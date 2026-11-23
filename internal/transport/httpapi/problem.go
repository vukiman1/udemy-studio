package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

const (
	problemContentType         = "application/problem+json"
	downloadsUnavailableDetail = "downloads are not configured: start the server with --python and --python-repo"
)

func problem(status int, detail string) gen.Problem {
	return gen.Problem{Title: http.StatusText(status), Status: status, Detail: &detail}
}

func notFound(detail string) gen.NotFoundApplicationProblemPlusJSONResponse {
	return gen.NotFoundApplicationProblemPlusJSONResponse(problem(http.StatusNotFound, detail))
}

func downloadsUnavailable() gen.DownloadsUnavailableApplicationProblemPlusJSONResponse {
	return gen.DownloadsUnavailableApplicationProblemPlusJSONResponse(problem(http.StatusServiceUnavailable, downloadsUnavailableDetail))
}

func writeProblem(writer http.ResponseWriter, status int, detail string) {
	writer.Header().Set("Content-Type", problemContentType)
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(problem(status, detail))
}

func strictErrorHandlers(logger *slog.Logger) gen.StrictHTTPServerOptions {
	return gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: writeBadRequestProblem,
		ResponseErrorHandlerFunc: func(writer http.ResponseWriter, request *http.Request, err error) {
			logger.Error("request failed", "method", request.Method, "path", request.URL.Path, "err", err)
			writeProblem(writer, http.StatusInternalServerError, "unexpected server error")
		},
	}
}

func writeBadRequestProblem(writer http.ResponseWriter, _ *http.Request, err error) {
	writeProblem(writer, http.StatusBadRequest, err.Error())
}
