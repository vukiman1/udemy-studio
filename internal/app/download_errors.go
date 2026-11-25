package app

import (
	"errors"

	"github.com/vukiman1/udemy-studio/internal/adapter/udemyapi"
	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/login"
)

func classifyDownloadError(err error) jobs.ErrorKind {
	var status *udemyapi.StatusError
	switch {
	case errors.As(err, &status) && status.IsAuthFailure(),
		errors.Is(err, login.ErrNotLoggedIn),
		errors.Is(err, login.ErrTokenExpired),
		errors.Is(err, login.ErrTokenRejected),
		errors.Is(err, login.ErrProfileNotFound):
		return jobs.ErrorKindLogin
	case errors.Is(err, udemyapi.ErrCourseNotFound):
		return jobs.ErrorKindCourseNotFound
	case errors.Is(err, udemyapi.ErrChallenged):
		return jobs.ErrorKindBlocked
	case errors.Is(err, login.ErrNoDiskAccess), errors.Is(err, login.ErrNoKeychainAccess):
		return jobs.ErrorKindChromeAccess
	default:
		return jobs.ErrorKindOther
	}
}
