package transcode

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
)

// Measured with libx265 -preset fast on Apple Silicon: 1080p lectures encode at 3.5–4x realtime.
const assumedRealtimeSpeed = 3.5

type Deps struct {
	Catalog *library.Catalog
	Prober  Prober
	Encoder Encoder
	Files   Files
	Skips   SkipStore
	Logger  *slog.Logger
}

// Service compresses one course at a time; the others wait in the order they were started.
type Service struct {
	deps          Deps
	wake          chan struct{}
	mutex         sync.Mutex
	runs          map[string]*Run
	plans         map[string]Plan
	queue         []string
	cancelRunning context.CancelFunc
	cache         mediaCache
	resurvey      chan struct{}
}

func NewService(deps Deps) *Service {
	return &Service{deps: deps, wake: make(chan struct{}, 1), runs: map[string]*Run{}, plans: map[string]Plan{}, cache: mediaCache{entries: map[string]cachedMedia{}}, resurvey: make(chan struct{}, 1)}
}

func (service *Service) Plan(ctx context.Context, courseID string) (Plan, error) {
	course, ok := service.deps.Catalog.Current().Course(courseID)
	if !ok {
		return Plan{}, fmt.Errorf("%w: %s", ErrCourseNotFound, courseID)
	}
	var plan Plan
	for _, chapter := range course.Chapters {
		for _, lecture := range chapter.Lectures {
			if lecture.VideoPath == "" {
				continue
			}
			if err := service.addToPlan(ctx, &plan, lecture); err != nil {
				return Plan{}, err
			}
		}
	}
	plan.EstimatedFor = time.Duration(float64(plan.TotalLength) / assumedRealtimeSpeed)
	return plan, nil
}

func (service *Service) addToPlan(ctx context.Context, plan *Plan, lecture library.Lecture) error {
	media, probeErr := service.probe(ctx, lecture.VideoPath)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	kind, err := service.classify(ctx, lecture.VideoPath, media, probeErr == nil)
	if err != nil {
		return err
	}
	switch kind {
	case videoUnreadable:
		service.deps.Logger.Warn("not compressing unreadable video", "path", lecture.VideoPath, "err", probeErr)
		plan.Unreadable++
	case videoHEVC:
		plan.AlreadyHEVC++
	case videoOtherCodec:
		plan.OtherCodec++
	case videoNotWorthIt:
		plan.NotWorthIt++
	case videoPending:
		plan.Candidates = append(plan.Candidates, Candidate{LectureTitle: lecture.Title, Path: lecture.VideoPath, Media: media})
		plan.TotalBytes += media.SizeBytes
		plan.TotalLength += media.Duration
	}
	return nil
}

type videoKind int

const (
	videoPending videoKind = iota
	videoHEVC
	videoOtherCodec
	videoNotWorthIt
	videoUnreadable
)

// classify tells what compressing a video would do; only an H.264 video not already found not worth it is pending.
func (service *Service) classify(ctx context.Context, path string, media Media, isReadable bool) (videoKind, error) {
	switch {
	case !isReadable:
		return videoUnreadable, nil
	case media.Codec == HEVCCodec:
		return videoHEVC, nil
	case media.Codec != H264Codec:
		return videoOtherCodec, nil
	}
	skip, found, err := service.deps.Skips.Find(ctx, path)
	if err != nil {
		return videoPending, fmt.Errorf("look up skipped %s: %w", path, err)
	}
	if found && skip.SizeBytes == media.SizeBytes {
		return videoNotWorthIt, nil
	}
	return videoPending, nil
}

// Summarize counts what is left to compress in a course without asking ffprobe; videos the survey has not read are Unchecked.
func (service *Service) Summarize(ctx context.Context, course library.Course) (Summary, error) {
	var summary Summary
	for _, chapter := range course.Chapters {
		for _, lecture := range chapter.Lectures {
			if lecture.VideoPath == "" {
				continue
			}
			summary.Videos++
			info, err := service.deps.Files.Stat(lecture.VideoPath)
			if err != nil {
				continue
			}
			cached, ok := service.cache.lookup(lecture.VideoPath, info)
			if !ok {
				summary.Unchecked++
				continue
			}
			kind, err := service.classify(ctx, lecture.VideoPath, cached.media, cached.probeErr == nil)
			if err != nil {
				return Summary{}, err
			}
			if kind == videoPending {
				summary.Pending++
			}
		}
	}
	return summary, nil
}

// Start plans the course again, since files may have changed since it was shown, and queues it.
func (service *Service) Start(ctx context.Context, courseID string) (Run, error) {
	if run, ok := service.Status(courseID); ok && run.State.IsActive() {
		return Run{}, fmt.Errorf("%w: %s", ErrAlreadyQueued, courseID)
	}
	plan, err := service.Plan(ctx, courseID)
	if err != nil {
		return Run{}, err
	}
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if run, ok := service.runs[courseID]; ok && run.State.IsActive() {
		return Run{}, fmt.Errorf("%w: %s", ErrAlreadyQueued, courseID)
	}
	run := &Run{CourseID: courseID, State: StateQueued, Total: len(plan.Candidates)}
	if run.Total == 0 {
		run.State = StateDone
	}
	service.runs[courseID] = run
	if run.State == StateQueued {
		service.plans[courseID] = plan
		service.queue = append(service.queue, courseID)
		service.signalWork()
	}
	return cloneRun(run), nil
}

func (service *Service) Cancel(courseID string) error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	run, ok := service.runs[courseID]
	if !ok || !run.State.IsActive() {
		return fmt.Errorf("%w: %s", ErrNotQueued, courseID)
	}
	if run.State == StateRunning {
		service.cancelRunning()
		return nil
	}
	service.queue = slices.DeleteFunc(service.queue, func(queued string) bool { return queued == courseID })
	delete(service.plans, courseID)
	run.State = StateCancelled
	return nil
}

func (service *Service) Status(courseID string) (Run, bool) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	run, ok := service.runs[courseID]
	if !ok {
		return Run{}, false
	}
	return cloneRun(run), true
}

func (service *Service) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		if courseID, plan, runCtx, ok := service.nextCourse(ctx); ok {
			service.compressCourse(runCtx, courseID, plan)
			continue
		}
		select {
		case <-ctx.Done():
		case <-service.wake:
		}
	}
	return nil
}

func (service *Service) nextCourse(ctx context.Context) (string, Plan, context.Context, bool) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if len(service.queue) == 0 {
		return "", Plan{}, nil, false
	}
	courseID := service.queue[0]
	service.queue = service.queue[1:]
	plan := service.plans[courseID]
	delete(service.plans, courseID)
	runCtx, cancel := context.WithCancel(ctx)
	service.cancelRunning = cancel
	service.runs[courseID].State = StateRunning
	return courseID, plan, runCtx, true
}

func (service *Service) compressCourse(ctx context.Context, courseID string, plan Plan) {
	for _, candidate := range plan.Candidates {
		if ctx.Err() != nil {
			break
		}
		service.update(courseID, func(run *Run) {
			run.Current = &Current{LectureTitle: candidate.LectureTitle, Length: candidate.Media.Duration}
		})
		result := service.compressFile(ctx, candidate, func(encoded time.Duration) {
			service.update(courseID, func(run *Run) { run.Current.Encoded = encoded })
		})
		service.update(courseID, func(run *Run) { result.applyTo(run, candidate) })
	}
	service.update(courseID, func(run *Run) {
		run.Current = nil
		run.State = StateDone
		if ctx.Err() != nil {
			run.State = StateCancelled
		}
	})
	service.mutex.Lock()
	service.cancelRunning()
	service.cancelRunning = nil
	service.mutex.Unlock()
}

func (service *Service) update(courseID string, change func(run *Run)) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	change(service.runs[courseID])
}

func (service *Service) signalWork() {
	select {
	case service.wake <- struct{}{}:
	default:
	}
}

func cloneRun(run *Run) Run {
	clone := *run
	clone.Skipped = slices.Clone(run.Skipped)
	clone.Failed = slices.Clone(run.Failed)
	if run.Current != nil {
		current := *run.Current
		clone.Current = &current
	}
	return clone
}
