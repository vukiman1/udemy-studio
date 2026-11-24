package httpapi

import (
	"net"
	"net/http"
	"slices"
	"strings"
)

const loopbackHostname = "localhost"

// RejectForeignHosts blocks DNS rebinding: a rebound page reaches us under its own domain name, never as an IP
// literal or as one of the public hostnames the server was told it is reached by, such as a Cloudflare Tunnel's.
func RejectForeignHosts(publicHosts []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !isTrustedHost(request.Host, publicHosts) {
			writeProblem(writer, http.StatusForbidden, "open the app by localhost, an IP address or its configured public host")
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func isTrustedHost(hostHeader string, publicHosts []string) bool {
	hostname := hostHeader
	if host, _, err := net.SplitHostPort(hostHeader); err == nil {
		hostname = host
	}
	hostname = strings.ToLower(hostname)
	return hostname == loopbackHostname || net.ParseIP(hostname) != nil || slices.Contains(publicHosts, hostname)
}
