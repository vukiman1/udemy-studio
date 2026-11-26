package publish

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path"
	"sync"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/objectstore"
)

const DefaultLimitBytes int64 = 9_500_000_000

var (
	ErrCourseNotFound = errors.New("no such course")
	ErrTooLarge       = errors.New("the course is larger than the cloud storage limit; compress it first")
	ErrBusy           = errors.New("a course is already being published")
	ErrNotRunning     = errors.New("no course is being published")
)

type State string

const (
	StateRunning   State = "running"
	StateDone      State = "done"
	StateFailed    State = "failed"
	StateCancelled State = "cancelled"
)

type BucketSource interface {
	Bucket(ctx context.Context) (objectstore.Bucket, error)
}

type Deps struct {
	Catalog    *library.Catalog
	Files      fs.FS
	Buckets    BucketSource
	LimitBytes int64
	Now        func() time.Time
	Logger     *slog.Logger
}

type Plan struct {
	Course     library.Course
	Files      []courseFile
	TotalBytes int64
	LimitBytes int64
	Published  *IndexCourse
}

func (plan Plan) Fits() bool {
	return plan.TotalBytes <= plan.LimitBytes
}

func (plan Plan) Replaces() bool {
	return plan.Published != nil && plan.Published.Dir != plan.Course.Dir
}

type Run struct {
	CourseID     string
	Title        string
	State        State
	TotalFiles   int
	SentFiles    int
	SkippedFiles int
	TotalBytes   int64
	SentBytes    int64
	CurrentFile  string
	Failure      string
}

type pushRequest struct {
	run    *Run
	plan   Plan
	bucket objectstore.Bucket
}

type Service struct {
	deps          Deps
	requests      chan pushRequest
	mutex         sync.Mutex
	run           *Run
	cancelRunning context.CancelFunc
}

func NewService(deps Deps) *Service {
	return &Service{deps: deps, requests: make(chan pushRequest, 1)}
}

func (service *Service) Plan(ctx context.Context, courseID string) (Plan, error) {
	plan, _, err := service.plan(ctx, courseID)
	return plan, err
}

func (service *Service) Published(ctx context.Context) (Index, bool, error) {
	bucket, err := service.deps.Buckets.Bucket(ctx)
	if err != nil {
		return Index{}, false, err
	}
	return ReadIndex(ctx, bucket)
}

func (service *Service) Start(ctx context.Context, courseID string) (Run, error) {
	plan, bucket, err := service.plan(ctx, courseID)
	if err != nil {
		return Run{}, err
	}
	if !plan.Fits() {
		return Run{}, fmt.Errorf("%w: %d bytes, limit %d", ErrTooLarge, plan.TotalBytes, plan.LimitBytes)
	}
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.run != nil && service.run.State == StateRunning {
		return Run{}, fmt.Errorf("%w: %s", ErrBusy, service.run.Title)
	}
	service.run = &Run{CourseID: courseID, Title: plan.Course.Title, State: StateRunning, TotalFiles: len(plan.Files), TotalBytes: plan.TotalBytes}
	service.requests <- pushRequest{run: service.run, plan: plan, bucket: bucket}
	return *service.run, nil
}

func (service *Service) Cancel() error {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.run == nil || service.run.State != StateRunning {
		return ErrNotRunning
	}
	if service.cancelRunning != nil {
		service.cancelRunning()
		return nil
	}
	service.run.State = StateCancelled
	return nil
}

func (service *Service) Status() (Run, bool) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.run == nil {
		return Run{}, false
	}
	return *service.run, true
}

func (service *Service) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case request := <-service.requests:
			service.push(ctx, request)
		}
	}
}

func (service *Service) plan(ctx context.Context, courseID string) (Plan, objectstore.Bucket, error) {
	course, found := service.deps.Catalog.Current().Course(courseID)
	if !found {
		return Plan{}, nil, fmt.Errorf("%w: %s", ErrCourseNotFound, courseID)
	}
	files, total, err := listCourseFiles(service.deps.Files, course.Dir)
	if err != nil {
		return Plan{}, nil, err
	}
	bucket, err := service.deps.Buckets.Bucket(ctx)
	if err != nil {
		return Plan{}, nil, err
	}
	plan := Plan{Course: course, Files: files, TotalBytes: total, LimitBytes: service.deps.LimitBytes}
	index, found, err := ReadIndex(ctx, bucket)
	if err != nil {
		return Plan{}, nil, err
	}
	if found {
		plan.Published = &index.Course
	}
	return plan, bucket, nil
}

func (service *Service) push(ctx context.Context, request pushRequest) {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	if !service.begin(request.run, cancel) {
		return
	}
	err := service.upload(runCtx, request)
	service.finish(runCtx, err)
}

func (service *Service) begin(run *Run, cancel context.CancelFunc) bool {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.run != run || run.State != StateRunning {
		return false
	}
	service.cancelRunning = cancel
	return true
}

func (service *Service) finish(ctx context.Context, err error) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.cancelRunning = nil
	service.run.CurrentFile = ""
	switch {
	case err == nil:
		service.run.State = StateDone
	case ctx.Err() != nil:
		service.run.State = StateCancelled
	default:
		service.deps.Logger.Error("publishing a course failed", "course", service.run.CourseID, "err", err)
		service.run.State = StateFailed
		service.run.Failure = err.Error()
	}
}

func (service *Service) upload(ctx context.Context, request pushRequest) error {
	plan, bucket := request.plan, request.bucket
	if plan.Replaces() {
		if err := bucket.Delete(ctx, objectstore.IndexKey); err != nil {
			return err
		}
	}
	remote, err := service.removeOthers(ctx, bucket, plan)
	if err != nil {
		return err
	}
	for _, file := range plan.Files {
		if err := service.send(ctx, bucket, plan.Course.Dir, file, remote); err != nil {
			return err
		}
	}
	return writeIndex(ctx, bucket, service.index(plan))
}

func (service *Service) removeOthers(ctx context.Context, bucket objectstore.Bucket, plan Plan) (map[string]objectstore.Object, error) {
	listed, err := bucket.List(ctx, CoursesPrefix)
	if err != nil {
		return nil, err
	}
	wanted := make(map[string]bool, len(plan.Files))
	for _, file := range plan.Files {
		wanted[ObjectKey(plan.Course.Dir, file.Path)] = true
	}
	kept := make(map[string]objectstore.Object, len(listed))
	for _, object := range listed {
		if wanted[object.Key] {
			kept[object.Key] = object
			continue
		}
		if err := bucket.Delete(ctx, object.Key); err != nil {
			return nil, err
		}
	}
	return kept, nil
}

func (service *Service) send(ctx context.Context, bucket objectstore.Bucket, courseDir string, file courseFile, remote map[string]objectstore.Object) error {
	key := ObjectKey(courseDir, file.Path)
	if existing, found := remote[key]; found && isUpToDate(existing, file) {
		service.update(func(run *Run) { run.SkippedFiles++; run.SentBytes += file.SizeBytes })
		return nil
	}
	service.update(func(run *Run) { run.CurrentFile = file.Path })
	if err := uploadFile(ctx, bucket, service.deps.Files, path.Join(courseDir, file.Path), key, file.SizeBytes); err != nil {
		return err
	}
	service.update(func(run *Run) { run.SentFiles++; run.SentBytes += file.SizeBytes })
	return nil
}

func isUpToDate(object objectstore.Object, file courseFile) bool {
	return object.SizeBytes == file.SizeBytes && !object.ModifiedAt.Before(file.ModifiedAt)
}

func uploadFile(ctx context.Context, bucket objectstore.Bucket, files fs.FS, localPath, key string, size int64) error {
	opened, err := files.Open(localPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", localPath, err)
	}
	defer func() { _ = opened.Close() }()
	body, ok := opened.(io.ReadSeeker)
	if !ok {
		return fmt.Errorf("open %s: the file cannot be re-read for signing", localPath)
	}
	return bucket.Put(ctx, objectstore.Upload{Key: key, Body: body, SizeBytes: size, ContentType: contentType(localPath)})
}

func (service *Service) index(plan Plan) Index {
	files := make([]IndexFile, len(plan.Files))
	for position, file := range plan.Files {
		files[position] = IndexFile{Path: file.Path, SizeBytes: file.SizeBytes}
	}
	return Index{
		Format:      IndexFormat,
		PublishedAt: service.deps.Now().UTC(),
		Course:      IndexCourse{Dir: plan.Course.Dir, Title: plan.Course.Title, SizeBytes: plan.TotalBytes},
		Files:       files,
	}
}

func (service *Service) update(change func(run *Run)) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	change(service.run)
}
