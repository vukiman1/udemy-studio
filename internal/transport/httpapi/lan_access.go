package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	_ "embed"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strings"
)

const (
	lanCookieName      = "studio_lan"
	lanCookieMaxAge    = 180 * 24 * 60 * 60
	lanCodeQueryParam  = "code"
	lanCookieHMACLabel = "udemy-studio lan access"
)

var protectedDataPrefixes = []string{apiPrefix, mediaPrefix, articlesPrefix, attachmentsPrefix, healthPath}

//go:embed lan_login.html
var lanLoginPage []byte

func RequireLANAccess(accessCode string, next http.Handler) http.Handler {
	expectedCookie := lanCookieValue(accessCode)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if isLoopback(request.RemoteAddr) || hasValidCookie(request, expectedCookie) {
			next.ServeHTTP(writer, request)
			return
		}
		if submitted := request.URL.Query().Get(lanCodeQueryParam); submitted != "" && sameSecret(submitted, accessCode) {
			grantAccess(writer, request, expectedCookie)
			return
		}
		if isDataRequest(request.URL.Path) {
			writeProblem(writer, http.StatusUnauthorized, "LAN access code required")
			return
		}
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = writer.Write(lanLoginPage)
	})
}

func grantAccess(writer http.ResponseWriter, request *http.Request, cookieValue string) {
	//nolint:gosec // reason: LAN mode is plain HTTP, so a Secure cookie would never be sent back
	http.SetCookie(writer, &http.Cookie{
		Name:     lanCookieName,
		Value:    cookieValue,
		Path:     "/",
		MaxAge:   lanCookieMaxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	//nolint:gosec // reason: sameOriginLocation keeps only a single-slash path, covered by TestLANAccessRedirectNeverLeavesTheServer
	http.Redirect(writer, request, sameOriginLocation(request.URL), http.StatusSeeOther)
}

func sameOriginLocation(requested *url.URL) string {
	query := requested.Query()
	query.Del(lanCodeQueryParam)
	location := url.URL{Path: "/" + strings.TrimLeft(requested.Path, "/"), RawQuery: query.Encode()}
	return location.String()
}

func lanCookieValue(accessCode string) string {
	mac := hmac.New(sha256.New, []byte(accessCode))
	mac.Write([]byte(lanCookieHMACLabel))
	return hex.EncodeToString(mac.Sum(nil))
}

func hasValidCookie(request *http.Request, expected string) bool {
	cookie, err := request.Cookie(lanCookieName)
	return err == nil && sameSecret(cookie.Value, expected)
}

func sameSecret(given, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(given), []byte(expected)) == 1
}

func isLoopback(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func isDataRequest(path string) bool {
	for _, prefix := range protectedDataPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
