package httpapi

import "net/http"

const cloudOnlyReadsDetail = "this copy only reads the course in cloud storage; download, translate, compress and publish on the library copy"

var libraryOnlyRoutes = []string{
	"/api/library/tools",
	"/api/downloads/",
	"/api/jobs",
	"/api/jobs/",
	"/api/events",
	"/api/ai/",
	"/api/publish",
	"/api/publish/",
	"/api/courses/{courseId}/compression",
	"/api/courses/{courseId}/compression/plan",
	"/api/courses/{courseId}/translation",
	"/api/courses/{courseId}/translation/plan",
	"/api/courses/{courseId}/publish",
	"/api/courses/{courseId}/publish/plan",
}

func RejectLibraryOnlyRoutes(next http.Handler) http.Handler {
	blocked := http.NewServeMux()
	for _, route := range libraryOnlyRoutes {
		blocked.HandleFunc(route, func(writer http.ResponseWriter, _ *http.Request) {
			writeProblem(writer, http.StatusForbidden, cloudOnlyReadsDetail)
		})
	}
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if _, pattern := blocked.Handler(request); pattern != "" {
			blocked.ServeHTTP(writer, request)
			return
		}
		next.ServeHTTP(writer, request)
	})
}
