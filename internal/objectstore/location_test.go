package objectstore

import (
	"errors"
	"testing"
)

func TestParseURLSplitsEndpointAndBucket(t *testing.T) {
	cases := map[string]Location{
		"https://abc123.r2.cloudflarestorage.com/studio-dev":     {Endpoint: "https://abc123.r2.cloudflarestorage.com", Bucket: "studio-dev"},
		"  https://abc123.eu.r2.cloudflarestorage.com/studio/  ": {Endpoint: "https://abc123.eu.r2.cloudflarestorage.com", Bucket: "studio"},
		"http://localhost:9000/studio-test":                      {Endpoint: "http://localhost:9000", Bucket: "studio-test"},
		"http://127.0.0.1:9000/studio-test":                      {Endpoint: "http://127.0.0.1:9000", Bucket: "studio-test"},
	}
	for raw, want := range cases {
		got, err := ParseURL(raw)
		if err != nil {
			t.Errorf("ParseURL(%q): %v", raw, err)
			continue
		}
		if got != want {
			t.Errorf("ParseURL(%q) = %+v, want %+v", raw, got, want)
		}
	}
}

func TestParseURLRejectsURLsThatCannotNameOneBucketSafely(t *testing.T) {
	rejected := []string{
		"",
		"abc123.r2.cloudflarestorage.com/studio-dev",
		"https://abc123.r2.cloudflarestorage.com",
		"https://abc123.r2.cloudflarestorage.com/",
		"https://abc123.r2.cloudflarestorage.com/studio-dev/courses",
		"https://abc123.r2.cloudflarestorage.com/Studio_Dev",
		"http://abc123.r2.cloudflarestorage.com/studio-dev",
		"https://user:pass@abc123.r2.cloudflarestorage.com/studio-dev",
		"https://abc123.r2.cloudflarestorage.com/studio-dev?x-id=1",
	}
	for _, raw := range rejected {
		if _, err := ParseURL(raw); !errors.Is(err, ErrInvalidURL) {
			t.Errorf("ParseURL(%q) error = %v, want ErrInvalidURL", raw, err)
		}
	}
}
