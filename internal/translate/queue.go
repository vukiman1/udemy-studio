package translate

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"slices"
	"sync"
	"unicode/utf8"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

const (
	SourceLanguage = "en"
	TargetLanguage = "vi"
	// Measured on a 303-cue lecture: prompt and cues cost about 0.55 input tokens per character of cue text,
	// and replies 1.1 (Claude Opus 5.5) to 2.0 (Claude Haiku 5.5) output tokens with thinking; the estimate takes the higher.
	inputTokensPerCharacter  = 0.55
	outputTokensPerCharacter = 2.0
	subtitleFileMode         = 0o644
	temporaryPrefix          = "."
	temporarySuffix          = ".translating"
)

var (
	ErrCourseNotFound = errors.New("course not found")
	ErrAlreadyQueued  = errors.New("course is already being translated")
	ErrNotQueued      = errors.New("course is not being translated")
)

// Files writes into the library; *os.Root satisfies it.
type Files interface {
	Stat(name string) (fs.FileInfo, error)
	WriteFile(name string, data []byte, perm fs.FileMode) error
	Rename(oldPath, newPath string) error
	Remove(path string) error
}

type QueueDeps struct {
	Catalog *library.Catalog
	Media   fs.FS
	Files   Files
	AI      AI
	// OnTranslated runs after a course gained subtitles, so the library and search pick them up.
	OnTranslated func(ctx context.Context) error
	Logger       *slog.Logger
}

type Candidate struct {
	LectureID    string
	LectureTitle string
	SourcePath   string
	TargetPath   string
	Characters   int
}

type Plan struct {
	Candidates        []Candidate
	AlreadyTranslated int
	Characters        int
	EstimatedUSD      float64
	// IsPriced is false when the first configured model's price is unknown, so EstimatedUSD means nothing.
	IsPriced bool
}

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateDone      State = "done"
	StateCancelled State = "cancelled"
	StateFailed    State = "failed"
)

func (state State) IsActive() bool { return state == StateQueued || state == StateRunning }

type Outcome struct {
	Path         string
	LectureTitle string
	Reason       string
}

type Run struct {
	CourseID       string
	State          State
	Total          int
	Finished       int
	CurrentLecture string
	CostUSD        float64
	Failed         []Outcome
	Error          string
}

// Queue translates one course at a time; the others wait in the order they were started.
type Queue struct {
	deps          QueueDeps
	wake          chan struct{}
	mutex         sync.Mutex
	runs          map[string]*Run
	waiting       map[string]Plan
	order         []string
	cancelRunning context.CancelFunc
}

func NewQueue(deps QueueDeps) *Queue {
	return &Queue{deps: deps, wake: make(chan struct{}, 1), runs: map[string]*Run{}, waiting: map[string]Plan{}}
}

func (queue *Queue) Plan(ctx context.Context, courseID string) (Plan, error) {
	course, ok := queue.deps.Catalog.Current().Course(courseID)
	if !ok {
		return Plan{}, fmt.Errorf("%w: %s", ErrCourseNotFound, courseID)
	}
	pending, alreadyTranslated := PendingLectures(course)
	plan := Plan{AlreadyTranslated: alreadyTranslated}
	for _, candidate := range pending {
		queue.addToPlan(&plan, candidate)
	}
	estimate, isPriced, err := queue.deps.AI.EstimateUSD(ctx, llm.Usage{
		InputTokens:  int64(float64(plan.Characters) * inputTokensPerCharacter),
		OutputTokens: int64(float64(plan.Characters) * outputTokensPerCharacter),
	})
	if err != nil {
		return Plan{}, err
	}
	plan.EstimatedUSD, plan.IsPriced = estimate, isPriced
	return plan, nil
}

func (queue *Queue) addToPlan(plan *Plan, candidate Candidate) {
	cues, err := subtitle.ParseSRTFile(queue.deps.Media, candidate.SourcePath)
	if err != nil {
		queue.deps.Logger.Warn("not translating unreadable subtitle", "path", candidate.SourcePath, "err", err)
		return
	}
	for _, cue := range cues {
		candidate.Characters += utf8.RuneCountInString(singleLine(cue.Text))
	}
	plan.Candidates = append(plan.Candidates, candidate)
	plan.Characters += candidate.Characters
}

// Start plans the course again, since files may have changed since it was shown, and queues it.
func (queue *Queue) Start(ctx context.Context, courseID string) (Run, error) {
	plan, err := queue.Plan(ctx, courseID)
	if err != nil {
		return Run{}, err
	}
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if run, ok := queue.runs[courseID]; ok && run.State.IsActive() {
		return Run{}, fmt.Errorf("%w: %s", ErrAlreadyQueued, courseID)
	}
	run := &Run{CourseID: courseID, State: StateQueued, Total: len(plan.Candidates)}
	queue.runs[courseID] = run
	if run.Total == 0 {
		run.State = StateDone
		return cloneRun(run), nil
	}
	queue.waiting[courseID] = plan
	queue.order = append(queue.order, courseID)
	select {
	case queue.wake <- struct{}{}:
	default:
	}
	return cloneRun(run), nil
}

func (queue *Queue) Cancel(courseID string) error {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	run, ok := queue.runs[courseID]
	if !ok || !run.State.IsActive() {
		return fmt.Errorf("%w: %s", ErrNotQueued, courseID)
	}
	if run.State == StateRunning {
		queue.cancelRunning()
		return nil
	}
	queue.order = slices.DeleteFunc(queue.order, func(queued string) bool { return queued == courseID })
	delete(queue.waiting, courseID)
	run.State = StateCancelled
	return nil
}

func (queue *Queue) Status(courseID string) (Run, bool) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	run, ok := queue.runs[courseID]
	if !ok {
		return Run{}, false
	}
	return cloneRun(run), true
}

func (queue *Queue) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		if courseID, plan, runCtx, ok := queue.next(ctx); ok {
			queue.translateCourse(runCtx, courseID, plan)
			continue
		}
		select {
		case <-ctx.Done():
		case <-queue.wake:
		}
	}
	return nil
}

func (queue *Queue) next(ctx context.Context) (string, Plan, context.Context, bool) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if len(queue.order) == 0 {
		return "", Plan{}, nil, false
	}
	courseID := queue.order[0]
	queue.order = queue.order[1:]
	plan := queue.waiting[courseID]
	delete(queue.waiting, courseID)
	runCtx, cancel := context.WithCancel(ctx)
	queue.cancelRunning = cancel
	queue.runs[courseID].State = StateRunning
	return courseID, plan, runCtx, true
}

func (queue *Queue) translateCourse(ctx context.Context, courseID string, plan Plan) {
	defer queue.finish(ctx, courseID)
	translatedAny := false
	// However the run ends, lectures it already wrote must reach the library and stop counting as pending.
	defer func() {
		if !translatedAny {
			return
		}
		if err := queue.deps.OnTranslated(context.WithoutCancel(ctx)); err != nil {
			queue.deps.Logger.Error("refresh the library after translating", "course", courseID, "err", err)
		}
	}()
	service := NewService(queue.deps.AI, queue.deps.Logger)
	for _, candidate := range plan.Candidates {
		if ctx.Err() != nil {
			return
		}
		queue.update(courseID, func(run *Run) { run.CurrentLecture = candidate.LectureTitle })
		spend, err := queue.translateLecture(ctx, service, candidate)
		queue.update(courseID, func(run *Run) { run.CostUSD += spend.CostUSD })
		switch {
		case ctx.Err() != nil:
			return
		case errors.Is(err, llm.ErrInvalidAPIKey), errors.Is(err, llm.ErrNotConfigured):
			queue.fail(courseID, err)
			return
		case err != nil:
			queue.update(courseID, func(run *Run) {
				run.Failed = append(run.Failed, Outcome{Path: candidate.SourcePath, LectureTitle: candidate.LectureTitle, Reason: err.Error()})
			})
		default:
			translatedAny = true
		}
		queue.update(courseID, func(run *Run) { run.Finished++ })
	}
}

// translateLecture writes the translation beside the original under a hidden name, and moves it into place only once complete.
func (queue *Queue) translateLecture(ctx context.Context, service *Service, candidate Candidate) (Spend, error) {
	// A run planned before a session's import, or a stopped run started again, finds some lectures already done.
	if _, err := queue.deps.Files.Stat(candidate.TargetPath); err == nil {
		return Spend{}, nil
	}
	cues, err := subtitle.ParseSRTFile(queue.deps.Media, candidate.SourcePath)
	if err != nil {
		return Spend{}, fmt.Errorf("read %s: %w", candidate.SourcePath, err)
	}
	translated, spend, err := service.TranslateCues(ctx, candidate.LectureTitle, cues)
	if err != nil {
		return spend, err
	}
	return spend, WriteSubtitle(queue.deps.Files, candidate.TargetPath, translated)
}

func (queue *Queue) finish(ctx context.Context, courseID string) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	run := queue.runs[courseID]
	run.CurrentLecture = ""
	switch {
	case run.State == StateFailed:
	case ctx.Err() != nil:
		run.State = StateCancelled
	default:
		run.State = StateDone
	}
	queue.cancelRunning()
	queue.cancelRunning = nil
}

func (queue *Queue) fail(courseID string, err error) {
	queue.update(courseID, func(run *Run) {
		run.State = StateFailed
		run.Error = err.Error()
	})
}

func (queue *Queue) update(courseID string, change func(run *Run)) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	change(queue.runs[courseID])
}

func cloneRun(run *Run) Run {
	clone := *run
	clone.Failed = slices.Clone(run.Failed)
	return clone
}
