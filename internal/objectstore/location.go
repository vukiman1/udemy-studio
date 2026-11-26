package objectstore

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

const (
	secureScheme   = "https"
	loopbackScheme = "http"
	localhostName  = "localhost"
)

var bucketNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$`)

type Location struct {
	Endpoint string
	Bucket   string
}

func (location Location) URL() string {
	return location.Endpoint + "/" + location.Bucket
}

func ParseURL(raw string) (Location, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return Location{}, fmt.Errorf("%w: %w", ErrInvalidURL, err)
	}
	if err := checkEndpoint(parsed); err != nil {
		return Location{}, err
	}
	bucket := strings.Trim(parsed.Path, "/")
	if bucket == "" || strings.Contains(bucket, "/") {
		return Location{}, fmt.Errorf("%w: end it with the bucket name, such as …/studio-dev", ErrInvalidURL)
	}
	if !bucketNamePattern.MatchString(bucket) {
		return Location{}, fmt.Errorf("%w: %q is not a bucket name", ErrInvalidURL, bucket)
	}
	return Location{Endpoint: parsed.Scheme + "://" + parsed.Host, Bucket: bucket}, nil
}

func checkEndpoint(parsed *url.URL) error {
	if parsed.Host == "" {
		return fmt.Errorf("%w: it needs a host, such as https://<account-id>.r2.cloudflarestorage.com/<bucket>", ErrInvalidURL)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return fmt.Errorf("%w: leave out user names, queries and fragments", ErrInvalidURL)
	}
	switch {
	case parsed.Scheme == secureScheme:
		return nil
	case parsed.Scheme == loopbackScheme && isLoopback(parsed.Hostname()):
		return nil
	default:
		return fmt.Errorf("%w: use https", ErrInvalidURL)
	}
}

func isLoopback(host string) bool {
	if host == localhostName {
		return true
	}
	address := net.ParseIP(host)
	return address != nil && address.IsLoopback()
}
