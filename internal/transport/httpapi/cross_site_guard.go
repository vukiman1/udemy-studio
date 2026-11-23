package httpapi

import (
	"net/http"
	"net/url"
)

var trustedFetchSites = map[string]bool{"same-origin": true, "none": true}

// RejectCrossSiteWrites stops other sites from using the browser to queue downloads or change progress:
// a text/plain POST needs no CORS preflight, so the server itself has to refuse it.
func RejectCrossSiteWrites(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if isSafeMethod(request.Method) || isSameSite(request) {
			next.ServeHTTP(writer, request)
			return
		}
		writeProblem(writer, http.StatusForbidden, "cross-site requests may not change data")
	})
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

func isSameSite(request *http.Request) bool {
	if site := request.Header.Get("Sec-Fetch-Site"); site != "" && !trustedFetchSites[site] {
		return false
	}
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}
	parsed, err := url.Parse(origin)
	return err == nil && parsed.Host == request.Host
}
