package main

import (
	"regexp"
	"strings"
)

const (
	redacted       = "REDACTED"
	fixturePortal  = "example-portal"
	signedParamRef = `${separator}${name}=` + redacted
)

var (
	signedURLParam    = regexp.MustCompile(`(?P<separator>[?&;])(?P<name>Expires|Signature|Key-Pair-Id|token|s|t)=[^&"\\\s]+`)
	jsonWebToken      = regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`)
	mediaLicenseToken = regexp.MustCompile(`"media_license_token": "[^"]*"`)
)

func sanitize(text, portal string) string {
	text = signedURLParam.ReplaceAllString(text, signedParamRef)
	text = jsonWebToken.ReplaceAllString(text, redacted)
	text = mediaLicenseToken.ReplaceAllString(text, `"media_license_token": "`+redacted+`"`)
	return strings.ReplaceAll(text, portal, fixturePortal)
}
