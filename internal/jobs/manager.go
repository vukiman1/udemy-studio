package jobs

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	jobIDBytes         = 8
	subscriberBuffer   = 32
	interruptedMessage = "interrupted because the server stopped"
)

type Downloader interface {
	Download(ctx context.Context, job Job, token secret.Secret, onEvent func(Event)) error
}

type TokenSource interface {
	Token(ctx context.Context, host string) (secret.Secret, error)
}

type Store interface {
	Save(ctx context.Context, job Job) error
	List(ctx context.Context) ([]Job, error)
}

type Deps struct {
	Downloader  Downloader
	Tokens      TokenSource
	Store       Store
	Logger      *slog.Logger
	Now         func() time.Time
	OnSucceeded func(ctx context.Context, job Job)
}

type Manager struct {
	deps        Deps
	wake        chan struct{}
	mutex       sync.Mutex
	jobs        map[string]*Job
	cancels     map[string]context.CancelFunc
	subscribers map[chan Job]struct{}
}

func NewManager(deps Deps) *Manager {
	return &Manager{
		deps:        deps,
		wake:        make(chan struct{}, 1),
		jobs:        map[string]*Job{},
		cancels:     map[string]context.CancelFunc{},
		subscribers: map[chan Job]struct{}{},
	}
}

func (manager *Manager) Restore(ctx context.Context) error {
	stored, err := manager.deps.Store.List(ctx)
	if err != nil {
		return fmt.Errorf("load jobs: %w", err)
	}
	slices.SortFunc(stored, func(left, right Job) int { return left.CreatedAt.Compare(right.CreatedAt) })
	for _, job := range stored {
		switch job.Status {
		case StatusRunning:
			job.Status, job.Error = StatusFailed, interruptedMessage
			job.FinishedAt = manager.now()
			job.Progress.CurrentLecture = ""
			if err := manager.deps.Store.Save(ctx, job); err != nil {
				return fmt.Errorf("mark job %s interrupted: %w", job.ID, err)
			}
		case StatusQueued, StatusSucceeded, StatusFailed, StatusCancelled:
		}
		manager.jobs[job.ID] = &job
	}
	manager.signalWork()
	return nil
}

func (manager *Manager) Submit(ctx context.Context, request Request) (Job, error) {
	if err := request.validate(); err != nil {
		return Job{}, err
	}
	id, err := newJobID()
	if err != nil {
		return Job{}, err
	}
	job := Job{ID: id, Request: request, Status: StatusQueued, CreatedAt: *manager.now()}
	if err := manager.deps.Store.Save(ctx, job); err != nil {
		return Job{}, fmt.Errorf("save job %s: %w", id, err)
	}
	tracked := job
	manager.mutex.Lock()
	manager.jobs[id] = &tracked
	manager.mutex.Unlock()
	manager.publish(job)
	manager.signalWork()
	return job, nil
}

func (manager *Manager) Cancel(ctx context.Context, id string) error {
	manager.mutex.Lock()
	job, ok := manager.jobs[id]
	if !ok {
		manager.mutex.Unlock()
		return fmt.Errorf("%w: %s", ErrJobNotFound, id)
	}
	if job.Status.IsFinished() {
		manager.mutex.Unlock()
		return fmt.Errorf("%w: %s is %s", ErrJobFinished, id, job.Status)
	}
	if cancel, running := manager.cancels[id]; running {
		manager.mutex.Unlock()
		cancel()
		return nil
	}
	job.Status, job.FinishedAt = StatusCancelled, manager.now()
	snapshot := cloneJob(job)
	manager.mutex.Unlock()
	return manager.persistAndPublish(ctx, snapshot)
}

func (manager *Manager) List() []Job {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	jobs := make([]Job, 0, len(manager.jobs))
	for _, job := range manager.jobs {
		jobs = append(jobs, cloneJob(job))
	}
	slices.SortFunc(jobs, func(left, right Job) int { return cmp.Compare(right.CreatedAt.UnixNano(), left.CreatedAt.UnixNano()) })
	return jobs
}

func (manager *Manager) Subscribe() (<-chan Job, func()) {
	updates := make(chan Job, subscriberBuffer)
	manager.mutex.Lock()
	manager.subscribers[updates] = struct{}{}
	manager.mutex.Unlock()
	return updates, func() {
		manager.mutex.Lock()
		delete(manager.subscribers, updates)
		manager.mutex.Unlock()
	}
}

func (manager *Manager) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		if id, ok := manager.nextQueued(); ok {
			manager.run(ctx, id)
			continue
		}
		select {
		case <-ctx.Done():
		case <-manager.wake:
		}
	}
	return nil
}

func (manager *Manager) nextQueued() (string, bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	var oldest *Job
	for _, job := range manager.jobs {
		if job.Status == StatusQueued && (oldest == nil || job.CreatedAt.Before(oldest.CreatedAt)) {
			oldest = job
		}
	}
	if oldest == nil {
		return "", false
	}
	return oldest.ID, true
}

func (manager *Manager) signalWork() {
	select {
	case manager.wake <- struct{}{}:
	default:
	}
}

func (manager *Manager) run(ctx context.Context, id string) {
	jobCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	manager.mutex.Lock()
	job, ok := manager.jobs[id]
	if !ok || job.Status != StatusQueued {
		manager.mutex.Unlock()
		return
	}
	job.Status, job.StartedAt = StatusRunning, manager.now()
	manager.cancels[id] = cancel
	started := cloneJob(job)
	manager.mutex.Unlock()
	manager.persistAndLog(ctx, started)

	err := manager.download(jobCtx, started)

	manager.mutex.Lock()
	delete(manager.cancels, id)
	job.FinishedAt = manager.now()
	job.Progress.CurrentLecture = ""
	switch {
	case err == nil:
		job.Status = StatusSucceeded
	case ctx.Err() != nil:
		job.Status, job.Error = StatusFailed, interruptedMessage
	case jobCtx.Err() != nil:
		job.Status = StatusCancelled
	default:
		job.Status, job.Error = StatusFailed, err.Error()
	}
	finished := cloneJob(job)
	manager.mutex.Unlock()
	manager.persistAndLog(context.WithoutCancel(ctx), finished)
	if finished.Status == StatusSucceeded && manager.deps.OnSucceeded != nil {
		manager.deps.OnSucceeded(ctx, finished)
	}
}

func (manager *Manager) download(ctx context.Context, job Job) error {
	token, err := manager.deps.Tokens.Token(ctx, job.Request.Host())
	if err != nil {
		return err
	}
	return manager.deps.Downloader.Download(ctx, job, token, func(event Event) {
		manager.mutex.Lock()
		current := manager.jobs[job.ID]
		current.apply(event)
		snapshot := cloneJob(current)
		manager.mutex.Unlock()
		manager.persistAndLog(context.WithoutCancel(ctx), snapshot)
	})
}

func (manager *Manager) persistAndLog(ctx context.Context, job Job) {
	if err := manager.persistAndPublish(ctx, job); err != nil {
		manager.deps.Logger.Error("save job state", "job", job.ID, "status", job.Status, "err", err)
	}
}

func (manager *Manager) persistAndPublish(ctx context.Context, job Job) error {
	manager.publish(job)
	if err := manager.deps.Store.Save(ctx, job); err != nil {
		return fmt.Errorf("save job %s: %w", job.ID, err)
	}
	return nil
}

// A subscriber that cannot keep up misses intermediate updates; the next one carries the full job state.
func (manager *Manager) publish(job Job) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	for subscriber := range manager.subscribers {
		select {
		case subscriber <- job:
		default:
		}
	}
}

func (manager *Manager) now() *time.Time {
	moment := manager.deps.Now().UTC()
	return &moment
}

func cloneJob(job *Job) Job {
	clone := *job
	clone.Failures = slices.Clone(job.Failures)
	return clone
}

func newJobID() (string, error) {
	random := make([]byte, jobIDBytes)
	if _, err := rand.Read(random); err != nil {
		return "", errors.Join(errors.New("generate job id"), err)
	}
	return hex.EncodeToString(random), nil
}
