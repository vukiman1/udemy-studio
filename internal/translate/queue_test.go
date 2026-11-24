package translate

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

const (
	chapterDir     = "course/01 - Intro"
	queueWaitStep  = 5 * time.Millisecond
	queueWaitLimit = 3 * time.Second
)

type failingModel struct{ err error }

func (model failingModel) Complete(context.Context, llm.Prompt) (llm.Completion, error) {
	return llm.Completion{}, model.err
}

func (failingModel) EstimateUSD(context.Context, llm.Usage) (float64, error) {
	return 1, nil
}

func (failingModel) IsReady(context.Context) error { return nil }

func writeLecture(t *testing.T, root, name string, languages ...string) {
	t.Helper()
	var document bytes.Buffer
	if err := subtitle.WriteSRT(&document, lectureCues(40)); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, filepath.FromSlash(chapterDir))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".mp4"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, language := range languages {
		if err := os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s_%s.srt", name, language)), document.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func startQueue(t *testing.T, root string, ai AI) (*Queue, string) {
	t.Helper()
	media, err := os.OpenRoot(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = media.Close() })
	catalog, err := library.NewCatalog(media.FS())
	if err != nil {
		t.Fatal(err)
	}
	queue := NewQueue(QueueDeps{
		Catalog:      catalog,
		Media:        media.FS(),
		Files:        media,
		AI:           ai,
		OnTranslated: func(context.Context) error { return catalog.Rescan() },
		Logger:       slog.New(slog.DiscardHandler),
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = queue.Run(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	return queue, catalog.Current().Courses[0].ID
}

func finishedRun(t *testing.T, queue *Queue, courseID string) Run {
	t.Helper()
	for deadline := time.Now().Add(queueWaitLimit); time.Now().Before(deadline); time.Sleep(queueWaitStep) {
		if run, _ := queue.Status(courseID); !run.State.IsActive() {
			return run
		}
	}
	t.Fatalf("translation of %s did not finish", courseID)
	return Run{}
}

func TestTranslationWritesSubtitlesOnlyForLecturesThatPassEveryCheck(t *testing.T) {
	root := t.TempDir()
	writeLecture(t, root, "001 Welcome", SourceLanguage)
	writeLecture(t, root, "002 Done already", SourceLanguage, TargetLanguage)
	queue, courseID := startQueue(t, root, &scriptedModel{reply: faithful})

	plan, err := queue.Plan(t.Context(), courseID)
	if err != nil || len(plan.Candidates) != 1 || plan.AlreadyTranslated != 1 {
		t.Fatalf("plan = %+v, %v; want only the untranslated lecture", plan, err)
	}
	if _, err := queue.Start(t.Context(), courseID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if run := finishedRun(t, queue, courseID); run.State != StateDone || run.Finished != 1 || len(run.Failed) != 0 || run.CostUSD <= 0 {
		t.Errorf("run = %+v", run)
	}
	written, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(chapterDir), "001 Welcome_vi.srt")) //nolint:gosec // reason: a path inside the test's own temp dir
	if err != nil || !bytes.Contains(written, []byte(translatedPrefix)) {
		t.Errorf("translated subtitle = %q, %v", written, err)
	}
	leftovers, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(chapterDir), ".*"))
	if len(leftovers) != 0 {
		t.Errorf("temporary files left behind: %v", leftovers)
	}
}

func TestATranslationThatFailsItsChecksLeavesNoSubtitleBehind(t *testing.T) {
	root := t.TempDir()
	writeLecture(t, root, "001 Welcome", SourceLanguage)
	queue, courseID := startQueue(t, root, &scriptedModel{reply: shifted})
	if _, err := queue.Start(t.Context(), courseID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if run := finishedRun(t, queue, courseID); run.State != StateDone || len(run.Failed) != 1 {
		t.Errorf("run = %+v, want the lecture listed as failed", run)
	}
	entries, _ := os.ReadDir(filepath.Join(root, filepath.FromSlash(chapterDir)))
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".mp4" && entry.Name() != "001 Welcome_en.srt" {
			t.Errorf("unexpected file %s", entry.Name())
		}
	}
}

func TestARejectedKeyStopsTheWholeRun(t *testing.T) {
	root := t.TempDir()
	writeLecture(t, root, "001 Welcome", SourceLanguage)
	writeLecture(t, root, "002 Next", SourceLanguage)
	queue, courseID := startQueue(t, root, failingModel{err: fmt.Errorf("%w by the provider (HTTP 401)", llm.ErrInvalidAPIKey)})
	if _, err := queue.Start(t.Context(), courseID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	run := finishedRun(t, queue, courseID)
	if run.State != StateFailed || run.Finished != 0 || run.Error == "" {
		t.Errorf("run = %+v, want it failed before the second lecture", run)
	}
}

// keyRevokedAfter answers its first calls, then reports the key rejected, as when it is revoked mid-run.
type keyRevokedAfter struct {
	scriptedModel
	answers int
}

func (model *keyRevokedAfter) Complete(ctx context.Context, prompt llm.Prompt) (llm.Completion, error) {
	if model.calls >= model.answers {
		return llm.Completion{}, fmt.Errorf("%w by the provider (HTTP 401)", llm.ErrInvalidAPIKey)
	}
	return model.scriptedModel.Complete(ctx, prompt)
}

func TestLecturesARunTranslatedBeforeItStoppedNoLongerCountAsPending(t *testing.T) {
	root := t.TempDir()
	writeLecture(t, root, "001 Welcome", SourceLanguage)
	writeLecture(t, root, "002 Next", SourceLanguage)
	queue, courseID := startQueue(t, root, &keyRevokedAfter{scriptedModel: scriptedModel{reply: faithful}, answers: 1})
	if _, err := queue.Start(t.Context(), courseID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if run := finishedRun(t, queue, courseID); run.State != StateFailed {
		t.Fatalf("run = %+v, want it stopped by the rejected key", run)
	}
	if plan, err := queue.Plan(t.Context(), courseID); err != nil || len(plan.Candidates) != 1 || plan.AlreadyTranslated != 1 {
		t.Errorf("plan after the stopped run = %+v, %v; want the translated lecture no longer pending", plan, err)
	}
}

func TestALectureTranslatedMeanwhileIsNeitherPaidForNorOverwritten(t *testing.T) {
	root := t.TempDir()
	writeLecture(t, root, "001 Welcome", SourceLanguage)
	writeLecture(t, root, "002 Next", SourceLanguage)
	imported := filepath.Join(root, filepath.FromSlash(chapterDir), "002 Next_vi.srt")
	model := &scriptedModel{reply: func(call int, lines [][2]string) string {
		if call == 1 {
			if err := os.WriteFile(imported, []byte("imported\n"), 0o600); err != nil {
				t.Error(err)
			}
		}
		return faithful(call, lines)
	}}
	queue, courseID := startQueue(t, root, model)
	if _, err := queue.Start(t.Context(), courseID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	finishedRun(t, queue, courseID)
	kept, err := os.ReadFile(imported) //nolint:gosec // reason: a path inside the test's own temp dir
	if err != nil || string(kept) != "imported\n" || model.calls != 1 {
		t.Errorf("imported subtitle = %q, %v after %d model calls; want it kept and not translated again", kept, err, model.calls)
	}
}

// quotaSpent never answers, as a provider does once its quota is used up.
type quotaSpent struct{ failingModel }

func (*quotaSpent) Complete(context.Context, llm.Prompt) (llm.Completion, error) {
	return llm.Completion{}, fmt.Errorf("%w: HTTP 429", llm.ErrQuotaExhausted)
}

func TestASpentQuotaStopsTheRunInsteadOfFailingEveryLecture(t *testing.T) {
	root := t.TempDir()
	writeLecture(t, root, "001 Welcome", SourceLanguage)
	writeLecture(t, root, "002 Next", SourceLanguage)
	queue, courseID := startQueue(t, root, &quotaSpent{})
	if _, err := queue.Start(t.Context(), courseID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if run := finishedRun(t, queue, courseID); run.State != StateFailed || run.Finished != 0 || len(run.Failed) != 0 {
		t.Errorf("run = %+v; want it stopped at the first lecture, not each lecture failed", run)
	}
}
