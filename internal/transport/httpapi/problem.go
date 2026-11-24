package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

const (
	problemContentType           = "application/problem+json"
	downloadsUnavailableDetail   = "downloads are not configured: start the server with --downloader go, or with --python and --python-repo"
	compressionUnavailableDetail = "compression needs ffmpeg and ffprobe; install them or pass their paths with --ffmpeg and --ffprobe"
)

func problem(status int, detail string) gen.Problem {
	return gen.Problem{Title: http.StatusText(status), Status: status, Detail: &detail}
}

func badRequest(err error) gen.BadRequestApplicationProblemPlusJSONResponse {
	return gen.BadRequestApplicationProblemPlusJSONResponse(problem(http.StatusBadRequest, err.Error()))
}

func notFound(detail string) gen.NotFoundApplicationProblemPlusJSONResponse {
	return gen.NotFoundApplicationProblemPlusJSONResponse(problem(http.StatusNotFound, detail))
}

func downloadsUnavailable() gen.DownloadsUnavailableApplicationProblemPlusJSONResponse {
	return gen.DownloadsUnavailableApplicationProblemPlusJSONResponse(problem(http.StatusServiceUnavailable, downloadsUnavailableDetail))
}

func compressionUnavailable() gen.CompressionUnavailableApplicationProblemPlusJSONResponse {
	return gen.CompressionUnavailableApplicationProblemPlusJSONResponse(problem(http.StatusServiceUnavailable, compressionUnavailableDetail))
}

func conflict(err error) gen.ConflictApplicationProblemPlusJSONResponse {
	return gen.ConflictApplicationProblemPlusJSONResponse(problem(http.StatusConflict, err.Error()))
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
