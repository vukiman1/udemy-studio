package httpapi

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

const testCourseURL = "https://acme.udemy.com/course/backend/"

type fakeCookieStore struct{}

func (fakeCookieStore) Profiles(context.Context) ([]login.Profile, error) {
	return []login.Profile{{Dir: "Profile 4", Name: "work", UdemyHosts: []string{"acme.udemy.com"}}, {Dir: "Default", Name: "home"}}, nil
}

func (store fakeCookieStore) Profile(ctx context.Context, profileDir string) (login.Profile, error) {
	profiles, _ := store.Profiles(ctx)
	for _, profile := range profiles {
		if profile.Dir == profileDir {
			return profile, nil
		}
	}
	return login.Profile{}, login.ErrProfileNotFound
}

func (fakeCookieStore) AccessToken(_ context.Context, profileDir, host string) (login.Token, error) {
	if profileDir != "Profile 4" || host != "acme.udemy.com" {
		return login.Token{}, login.ErrNotLoggedIn
	}
	return login.Token{Value: secret.New("token"), ExpiresAt: time.Now().Add(time.Hour)}, nil
}

type fakeSettings struct {
	selected string
	err      error
}

func (settings *fakeSettings) SelectedProfile(context.Context) (string, bool, error) {
	return settings.selected, settings.selected != "", settings.err
}

func (settings *fakeSettings) SaveSelectedProfile(_ context.Context, profileDir string) error {
	settings.selected = profileDir
	return nil
}

type memoryJobStore struct{}

func (memoryJobStore) Save(context.Context, jobs.Job) error     { return nil }
func (memoryJobStore) List(context.Context) ([]jobs.Job, error) { return nil, nil }

type blockingDownloader struct{}

func (blockingDownloader) Download(ctx context.Context, _ jobs.Job, _ secret.Secret, onEvent func(jobs.Event)) error {
	onEvent(jobs.Event{Kind: jobs.EventCourseStarted, CourseTitle: "Backend", TotalLectures: 3})
	<-ctx.Done()
	return ctx.Err()
}

func newDownloadsApp(t *testing.T) testApp {
	t.Helper()
	loginService := login.NewService(fakeCookieStore{}, &fakeSettings{}, time.Now)
	manager := jobs.NewManager(jobs.Deps{
		Downloader: blockingDownloader{}, Tokens: loginService, Store: memoryJobStore{},
		Logger: slog.New(slog.DiscardHandler), Now: time.Now,
	})
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		_ = manager.Run(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})
	return newTestAppWithDeps(t, libraryFixture, func(deps *Deps) {
		deps.Login, deps.Jobs = loginService, manager
	})
}

func TestDownloadEndpointsExplainWhenDownloadsAreOff(t *testing.T) {
	app := newTestApp(t)
	requests := map[string][2]string{
		"profiles": {http.MethodGet, "/api/downloads/profiles"},
		"status":   {http.MethodGet, "/api/downloads/status"},
		"jobs":     {http.MethodGet, "/api/jobs"},
		"cancel":   {http.MethodDelete, "/api/jobs/x"},
		"events":   {http.MethodGet, "/api/events"},
	}
	for name, request := range requests {
		t.Run(name, func(t *testing.T) {
			result := app.do(t, request[0], request[1], nil)
			assertProblem(t, result, http.StatusServiceUnavailable)
			if !strings.Contains(string(result.body), "--python") {
				t.Errorf("problem does not say how to enable downloads: %s", result.body)
			}
		})
	}
}

func TestSelectingAProfileShowsItsLoginStatus(t *testing.T) {
	app := newDownloadsApp(t)
	status := decode[gen.LoginStatus](t, app.do(t, http.MethodGet, "/api/downloads/status", nil))
	if status.Problem == nil || status.Profile != nil {
		t.Fatalf("status before selecting = %+v, want a problem and no profile", status)
	}
	assertProblem(t, app.do(t, http.MethodPut, "/api/downloads/profile", gen.ProfileSelection{Dir: "Profile 99"}), http.StatusNotFound)
	if result := app.do(t, http.MethodPut, "/api/downloads/profile", gen.ProfileSelection{Dir: "Profile 4"}); result.status != http.StatusNoContent {
		t.Fatalf("select status = %d %s", result.status, result.body)
	}

	status = decode[gen.LoginStatus](t, app.do(t, http.MethodGet, "/api/downloads/status", nil))
	if status.Problem != nil || status.Profile == nil || status.Profile.Dir != "Profile 4" || len(status.Hosts) != 1 || status.Hosts[0].IsExpired {
		t.Errorf("status after selecting = %+v", status)
	}
	profiles := decode[[]gen.ChromeProfile](t, app.do(t, http.MethodGet, "/api/downloads/profiles", nil))
	if len(profiles) != 2 || !profiles[0].Selected || profiles[1].Selected || len(profiles[1].UdemyHosts) != 0 {
		t.Errorf("profiles = %+v", profiles)
	}
}

func TestJobLifecycleOverHTTP(t *testing.T) {
	app := newDownloadsApp(t)
	app.do(t, http.MethodPut, "/api/downloads/profile", gen.ProfileSelection{Dir: "Profile 4"})

	assertProblem(t, app.do(t, http.MethodPost, "/api/jobs", gen.JobRequest{CourseUrl: "https://example.com/course/x/"}), http.StatusBadRequest)
	chapters := "4"
	created := app.do(t, http.MethodPost, "/api/jobs", gen.JobRequest{CourseUrl: testCourseURL, Chapters: &chapters})
	job := decode[gen.Job](t, created)
	if created.status != http.StatusAccepted || job.Status != gen.Queued || job.Chapters == nil || *job.Chapters != "4" {
		t.Fatalf("create = %d %+v", created.status, job)
	}

	waitForJob(t, app, job.Id, gen.Running)
	if result := app.do(t, http.MethodDelete, "/api/jobs/"+job.Id, nil); result.status != http.StatusNoContent {
		t.Fatalf("cancel = %d %s", result.status, result.body)
	}
	waitForJob(t, app, job.Id, gen.Cancelled)
	assertProblem(t, app.do(t, http.MethodDelete, "/api/jobs/"+job.Id, nil), http.StatusConflict)
	assertProblem(t, app.do(t, http.MethodDelete, "/api/jobs/missing", nil), http.StatusNotFound)
}

func TestEventStreamSendsJobsAsTheyChange(t *testing.T) {
	app := newDownloadsApp(t)
	app.do(t, http.MethodPut, "/api/downloads/profile", gen.ProfileSelection{Dir: "Profile 4"})
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, app.server.URL+"/api/events", nil)
	response, err := app.server.Client().Do(request)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}
	defer response.Body.Close()
	if response.Header.Get("Content-Type") != eventStreamContentType {
		t.Fatalf("content type = %q", response.Header.Get("Content-Type"))
	}

	created := decode[gen.Job](t, app.do(t, http.MethodPost, "/api/jobs", gen.JobRequest{CourseUrl: testCourseURL}))
	scanner := bufio.NewScanner(response.Body)
	for scanner.Scan() {
		data, ok := strings.CutPrefix(scanner.Text(), "data: ")
		if !ok {
			continue
		}
		var job gen.Job
		if err := json.Unmarshal([]byte(data), &job); err != nil {
			t.Fatalf("decode event %q: %v", data, err)
		}
		if job.Id == created.Id && job.Status == gen.Running {
			return
		}
	}
	t.Fatalf("stream ended without the running job: %v", scanner.Err())
}

func waitForJob(t *testing.T, app testApp, id string, want gen.JobStatus) gen.Job {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		for _, job := range decode[[]gen.Job](t, app.do(t, http.MethodGet, "/api/jobs", nil)) {
			if job.Id == id && job.Status == want {
				return job
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("job %s never became %s", id, want)
	return gen.Job{}
}

func TestLoginStatusRaisesFailuresTheLearnerCannotFix(t *testing.T) {
	loginService := login.NewService(fakeCookieStore{}, &fakeSettings{err: errors.New("database is locked")}, time.Now)
	app := newTestAppWithDeps(t, libraryFixture, func(deps *Deps) { deps.Login = loginService })
	assertProblem(t, app.do(t, http.MethodGet, "/api/downloads/status", nil), http.StatusInternalServerError)
}
