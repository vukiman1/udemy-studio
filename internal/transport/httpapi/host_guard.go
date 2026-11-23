package httpapi

import (
	"net"
	"net/http"
)

const loopbackHostname = "localhost"

// RejectForeignHosts blocks DNS rebinding: a rebound page reaches us under its own domain name, never as an IP literal.
func RejectForeignHosts(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !isTrustedHost(request.Host) {
			writeProblem(writer, http.StatusForbidden, "open the app by localhost or an IP address")
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func isTrustedHost(hostHeader string) bool {
	hostname := hostHeader
	if host, _, err := net.SplitHostPort(hostHeader); err == nil {
		hostname = host
	}
	return hostname == loopbackHostname || net.ParseIP(hostname) != nil
}
