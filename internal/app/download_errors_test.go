package app

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/adapter/udemyapi"
	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/login"
)

func TestDownloadErrorsAreSortedIntoWhatTheLearnerCanDo(t *testing.T) {
	cases := map[string]struct {
		err  error
		want jobs.ErrorKind
	}{
		"udemy refuses the token":     {fmt.Errorf("fetch curriculum: %w", &udemyapi.StatusError{Status: http.StatusUnauthorized}), jobs.ErrorKindLogin},
		"no login for the portal":     {fmt.Errorf("%w: paste a token", login.ErrNotLoggedIn), jobs.ErrorKindLogin},
		"chrome login expired":        {login.ErrTokenExpired, jobs.ErrorKindLogin},
		"course missing from account": {fmt.Errorf("%w (%q)", udemyapi.ErrCourseNotFound, "x"), jobs.ErrorKindCourseNotFound},
		"cloudflare challenge":        {fmt.Errorf("fetch: %w", udemyapi.ErrChallenged), jobs.ErrorKindBlocked},
		"keychain refused":            {login.ErrNoKeychainAccess, jobs.ErrorKindChromeAccess},
		"udemy server error":          {&udemyapi.StatusError{Status: http.StatusInternalServerError}, jobs.ErrorKindOther},
		"tool failed":                 {errors.New("yt-dlp exited with status 1"), jobs.ErrorKindOther},
	}
	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			if got := classifyDownloadError(testCase.err); got != testCase.want {
				t.Errorf("kind = %q, want %q", got, testCase.want)
			}
		})
	}
}
