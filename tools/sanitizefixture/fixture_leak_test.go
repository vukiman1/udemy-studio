package main

import (
	"io/fs"
	"os"
	"regexp"
	"testing"
)

const fixtureRoot = "../../testdata"

var (
	signedParamValue = regexp.MustCompile(`[?&;](Expires|Signature|Key-Pair-Id|token|s|t)=([^&"\\\s]*)`)
	portalHost       = regexp.MustCompile(`https?://([a-z0-9-]+)\.udemy\.com`)
	portalNameField  = regexp.MustCompile(`"portal_name": "([^"]*)"`)
)

var leakPatterns = map[string]*regexp.Regexp{
	"json web token": jsonWebToken,
	"license token":  regexp.MustCompile(`"media_license_token": "(?:[^R"]|R[^E"]|RE[^D"])`),
	"bearer header":  regexp.MustCompile(`(?i)bearer [A-Za-z0-9]`),
	"home directory": regexp.MustCompile(`/Users/[^/"]+/`),
}

var publicUdemyHosts = map[string]bool{"www": true, fixturePortal: true}

func TestFixturesContainNoSecrets(t *testing.T) {
	fixtures := os.DirFS(fixtureRoot)
	err := fs.WalkDir(fixtures, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		content, err := fs.ReadFile(fixtures, path)
		if err != nil {
			return err
		}
		if leaks := findLeaks(content); len(leaks) != 0 {
			t.Errorf("%s: %v", path, leaks)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", fixtureRoot, err)
	}
}

func findLeaks(content []byte) []string {
	var leaks []string
	for name, pattern := range leakPatterns {
		if pattern.Match(content) {
			leaks = append(leaks, name)
		}
	}
	for _, match := range signedParamValue.FindAllSubmatch(content, -1) {
		if string(match[2]) != redacted {
			leaks = append(leaks, "unredacted "+string(match[1])+" param")
		}
	}
	for _, match := range portalHost.FindAllSubmatch(content, -1) {
		if !publicUdemyHosts[string(match[1])] {
			leaks = append(leaks, "real portal host")
		}
	}
	for _, match := range portalNameField.FindAllSubmatch(content, -1) {
		if string(match[1]) != fixturePortal {
			leaks = append(leaks, "real portal_name")
		}
	}
	return leaks
}

func TestFindLeaksCatchesEachKind(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{"signature starting with R", `"https://vtt-c.udemycdn.com/a.vtt?Expires=REDACTED&Signature=Rk9vYmFy"`},
		{"short s param", `"https://x.udemycdn.com/a?s=abc"`},
		{"html escaped t param", `href=\"https://x.test/?a=1&amp;t=secret\"`},
		{"real portal host", `"https://acme-corp.udemy.com/course/x"`},
		{"real portal name", `"portal_name": "acme-corp"`},
		{"license token starting with R", `"media_license_token": "Rabc"`},
		{"json web token", `"eyJhbGciOi.eyJzdWIiOi.c2lnbmF0dXJl"`},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if leaks := findLeaks([]byte(testCase.content)); len(leaks) == 0 {
				t.Errorf("findLeaks(%s) found nothing", testCase.content)
			}
		})
	}
}

func TestFindLeaksAcceptsSanitizedOutput(t *testing.T) {
	raw := `"portal_name": "acme-corp", "url": "https://acme-corp.udemy.com/x.m3u8?provider=cf&token=abc&v=1", ` +
		`"vtt": "https://vtt-c.udemycdn.com/a.vtt?Expires=1&Signature=Rk9v&Key-Pair-Id=K1", "media_license_token": "opaque"`
	if leaks := findLeaks([]byte(sanitize(raw, "acme-corp"))); len(leaks) != 0 {
		t.Errorf("sanitized output still leaks: %v", leaks)
	}
}
