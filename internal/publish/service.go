package publish

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path"
	"slices"
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
	LimitBytes int64
	Published  *Index
}

func (plan Plan) Replaces() bool {
	return plan.Published != nil && plan.Published.Course.Dir != plan.Course.Dir
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
	run       *Run
	plan      Plan
	selection Selection
	bucket    objectstore.Bucket
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

func (service *Service) Start(ctx context.Context, courseID string, chapters []string) (Run, error) {
	plan, bucket, err := service.plan(ctx, courseID)
	if err != nil {
		return Run{}, err
	}
	selection, err := plan.Select(chapters)
	if err != nil {
		return Run{}, err
	}
	if selection.TotalBytes > plan.LimitBytes {
		return Run{}, fmt.Errorf("%w: %d bytes, limit %d", ErrTooLarge, selection.TotalBytes, plan.LimitBytes)
	}
	service.mutex.Lock()
	defer service.mutex.Unlock()
	if service.run != nil && service.run.State == StateRunning {
		return Run{}, fmt.Errorf("%w: %s", ErrBusy, service.run.Title)
	}
	service.run = &Run{CourseID: courseID, Title: plan.Course.Title, State: StateRunning, TotalFiles: len(selection.Files), TotalBytes: selection.TotalBytes}
	service.requests <- pushRequest{run: service.run, plan: plan, selection: selection, bucket: bucket}
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
	files, err := listCourseFiles(service.deps.Files, course.Dir)
	if err != nil {
		return Plan{}, nil, err
	}
	bucket, err := service.deps.Buckets.Bucket(ctx)
	if err != nil {
		return Plan{}, nil, err
	}
	plan := Plan{Course: course, Files: files, LimitBytes: service.deps.LimitBytes}
	index, found, err := ReadIndex(ctx, bucket)
	if err != nil {
		return Plan{}, nil, err
	}
	if found {
		plan.Published = &index
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
	plan, selection, bucket := request.plan, request.selection, request.bucket
	listed, err := bucket.List(ctx, CoursesPrefix)
	if err != nil {
		return err
	}
	kept, unwanted := splitByWanted(listed, plan.Course.Dir, selection.Files)
	if plan.Replaces() || listsAny(plan.Published, unwanted) {
		if err := bucket.Delete(ctx, objectstore.IndexKey); err != nil {
			return err
		}
	}
	for _, key := range unwanted {
		if err := bucket.Delete(ctx, key); err != nil {
			return err
		}
	}
	for _, file := range selection.Files {
		if err := service.send(ctx, bucket, plan.Course.Dir, file, kept); err != nil {
			return err
		}
	}
	return writeIndex(ctx, bucket, service.index(plan, selection))
}

func listsAny(index *Index, keys []string) bool {
	if index == nil {
		return false
	}
	listed := make(map[string]bool, len(index.Files))
	for _, file := range index.Files {
		listed[ObjectKey(index.Course.Dir, file.Path)] = true
	}
	return slices.ContainsFunc(keys, func(key string) bool { return listed[key] })
}

func splitByWanted(listed []objectstore.Object, courseDir string, files []courseFile) (map[string]objectstore.Object, []string) {
	wanted := make(map[string]bool, len(files))
	for _, file := range files {
		wanted[ObjectKey(courseDir, file.Path)] = true
	}
	kept := make(map[string]objectstore.Object, len(listed))
	var unwanted []string
	for _, object := range listed {
		if wanted[object.Key] {
			kept[object.Key] = object
			continue
		}
		unwanted = append(unwanted, object.Key)
	}
	return kept, unwanted
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

func (service *Service) index(plan Plan, selection Selection) Index {
	files := make([]IndexFile, len(selection.Files))
	for position, file := range selection.Files {
		files[position] = IndexFile{Path: file.Path, SizeBytes: file.SizeBytes}
	}
	return Index{
		Format:      IndexFormat,
		PublishedAt: service.deps.Now().UTC(),
		Course: IndexCourse{
			Dir:          plan.Course.Dir,
			Title:        plan.Course.Title,
			SizeBytes:    selection.TotalBytes,
			ChapterCount: len(plan.Chapters()),
		},
		Chapters: selection.Chapters,
		Files:    files,
	}
}

func (service *Service) update(change func(run *Run)) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	change(service.run)
}
