package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	http "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const (
	visitURLTemplate        = "https://%s.udemy.com/api-2.0/visits/current/?fields%%5Bvisit%%5D=@default,visitor,country&locale=en_US"
	courseSearchURLTemplate = "https://%s.udemy.com/api-2.0/users/me/subscribed-courses?fields[course]=id,url,title,published_title&page=1&page_size=500&search=%s"
	curriculumURLTemplate   = "https://%s.udemy.com/api-2.0/courses/%d/subscriber-curriculum-items/"

	chromeUserAgent    = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/154.0.0.0 Safari/537.36"
	requestTimeoutSecs = 120
	netscapeFieldCount = 7
	httpOnlyPrefix     = "#HttpOnly_"

	transportTLSClient = "tls-client"
	transportStdlib    = "stdlib"
)

var challengeMarkers = []string{"challenge-platform", "<title>Just a moment...</title>"}

var udemyAppHeaders = map[string]string{
	"accept-encoding":        "gzip",
	"x-mobile-visit-enabled": "true",
	"x-udemy-client-secret":  "f2lgDUDxjFiOlVHUpwQNFUfCQPyMO0tJQMaud53PF01UKueW8enYjeEYoyVeP0bb2XVEDkJ5GLJaVTfM5QgMVz6yyXyydZdA5QhzgvG9UmCPUYaCrIVf7VpmiilfbLJc",
	"x-udemy-client-id":      "TH96Ov3Ebo3OtgoSH5mOYzYolcowM3ycedWQDDce",
	"accept-language":        "en_US",
	"x-version-name":         "9.51.2",
	"x-client-name":          "Udemy-Android",
	"user-agent":             chromeUserAgent,
}

var curriculumParams = url.Values{
	"fields[lecture]":  {"title,object_index,created,asset,supplementary_assets,description,download_url"},
	"fields[quiz]":     {"title,object_index,type"},
	"fields[practice]": {"title,object_index"},
	"fields[chapter]":  {"title,object_index"},
	"fields[asset]":    {"title,filename,asset_type,status,is_external,media_license_token,course_is_drmed,media_sources,captions,slides,slide_urls,download_urls,external_url,stream_urls,@min,status,delayed_asset_message,processing_errors,body"},
	"caching_intent":   {"True"},
	"page_size":        {"200"},
}

var errChallenged = errors.New("cloudflare challenge in response")

type probeResult struct {
	status int
	body   []byte
}

type fetchFunc func(target string) (probeResult, error)

type courseSearchResponse struct {
	Results []struct {
		ID             int    `json:"id"`
		PublishedTitle string `json:"published_title"`
	} `json:"results"`
}

type curriculumResponse struct {
	Count   int               `json:"count"`
	Next    *string           `json:"next"`
	Results []json.RawMessage `json:"results"`
}

func main() {
	cookiesPath := flag.String("cookies", "", "Netscape cookies.txt exported from the browser")
	portal := flag.String("portal", "", "Udemy portal subdomain (required)")
	courseSlug := flag.String("course", "fundamentals-of-backend-communications-and-protocols", "course published_title")
	transport := flag.String("transport", transportTLSClient, "tls-client | stdlib")
	rounds := flag.Int("rounds", 20, "consecutive visit+curriculum rounds")
	pause := flag.Duration("pause", time.Second, "pause between rounds")
	flag.Parse()
	if *portal == "" {
		fmt.Fprintln(os.Stderr, "FAIL: -portal is required")
		os.Exit(1)
	}

	if err := run(*cookiesPath, *portal, *courseSlug, *transport, *rounds, *pause); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
}

func run(cookiesPath, portal, courseSlug, transport string, rounds int, pause time.Duration) error {
	cookies, err := loadNetscapeCookies(cookiesPath, time.Now())
	if err != nil {
		return err
	}
	fmt.Printf("transport=%s, loaded %d unexpired cookies\n", transport, len(cookies))

	fetch, err := newFetcher(transport, portal, cookies)
	if err != nil {
		return err
	}

	visitURL := fmt.Sprintf(visitURLTemplate, portal)
	firstVisit, err := fetch(visitURL)
	if err != nil {
		return fmt.Errorf("first visit: %w", err)
	}
	fmt.Printf("first visit: status=%d\n", firstVisit.status)

	courseID, err := findCourseID(fetch, portal, courseSlug)
	if err != nil {
		return err
	}
	fmt.Printf("course id resolved: %d\n", courseID)

	curriculumURL := fmt.Sprintf(curriculumURLTemplate, portal, courseID) + "?" + curriculumParams.Encode()
	for round := 1; round <= rounds; round++ {
		visit, err := fetch(visitURL)
		if err != nil {
			return fmt.Errorf("round %d visit: %w", round, err)
		}
		curriculum, err := fetch(curriculumURL)
		if err != nil {
			return fmt.Errorf("round %d curriculum: %w", round, err)
		}
		items, err := countCurriculumItems(curriculum.body)
		if err != nil {
			return fmt.Errorf("round %d curriculum: %w", round, err)
		}
		fmt.Printf("round %02d  visit=%d (%dB)  curriculum=%d (%dB, %d items)\n",
			round, visit.status, len(visit.body), curriculum.status, len(curriculum.body), items)
		time.Sleep(pause)
	}
	fmt.Printf("PASS: %d rounds (%d requests) without challenge\n", rounds, rounds*2)
	return nil
}

func newFetcher(transport, portal string, cookies []*http.Cookie) (fetchFunc, error) {
	switch transport {
	case transportTLSClient:
		return newTLSClientFetcher(portal, cookies)
	case transportStdlib:
		return newStdlibFetcher(cookies), nil
	default:
		return nil, fmt.Errorf("unknown transport %q", transport)
	}
}

func newTLSClientFetcher(portal string, cookies []*http.Cookie) (fetchFunc, error) {
	client, err := tlsclient.NewHttpClient(tlsclient.NewNoopLogger(),
		tlsclient.WithTimeoutSeconds(requestTimeoutSecs),
		tlsclient.WithClientProfile(profiles.Chrome_154),
		tlsclient.WithCookieJar(tlsclient.NewCookieJar()),
	)
	if err != nil {
		return nil, fmt.Errorf("create tls client: %w", err)
	}
	portalURL, err := url.Parse(fmt.Sprintf("https://%s.udemy.com/", portal))
	if err != nil {
		return nil, fmt.Errorf("parse portal url: %w", err)
	}
	client.SetCookies(portalURL, cookies)

	return func(target string) (probeResult, error) {
		request, err := http.NewRequest(http.MethodGet, target, nil)
		if err != nil {
			return probeResult{}, fmt.Errorf("build request: %w", err)
		}
		for name, value := range udemyAppHeaders {
			request.Header.Set(name, value)
		}
		response, err := client.Do(request)
		if err != nil {
			return probeResult{}, fmt.Errorf("do request: %w", err)
		}
		defer response.Body.Close()
		decoded := http.DecompressBody(response)
		defer decoded.Close()
		return checkResponse(response.StatusCode, decoded)
	}, nil
}

func newStdlibFetcher(cookies []*http.Cookie) fetchFunc {
	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		pairs = append(pairs, cookie.Name+"="+cookie.Value)
	}
	cookieHeader := strings.Join(pairs, "; ")
	client := &stdhttp.Client{Timeout: requestTimeoutSecs * time.Second}

	return func(target string) (probeResult, error) {
		request, err := stdhttp.NewRequest(stdhttp.MethodGet, target, nil)
		if err != nil {
			return probeResult{}, fmt.Errorf("build request: %w", err)
		}
		for name, value := range udemyAppHeaders {
			request.Header.Set(name, value)
		}
		request.Header.Set("cookie", cookieHeader)
		response, err := client.Do(request)
		if err != nil {
			return probeResult{}, fmt.Errorf("do request: %w", err)
		}
		defer response.Body.Close()
		body := io.Reader(response.Body)
		if response.Header.Get("content-encoding") == "gzip" {
			gzipReader, err := gzip.NewReader(response.Body)
			if err != nil {
				return probeResult{}, fmt.Errorf("open gzip body (status %d): %w", response.StatusCode, err)
			}
			defer gzipReader.Close()
			body = gzipReader
		}
		return checkResponse(response.StatusCode, body)
	}
}

func checkResponse(status int, body io.Reader) (probeResult, error) {
	content, err := io.ReadAll(body)
	if err != nil {
		return probeResult{}, fmt.Errorf("read body: %w", err)
	}
	if isChallenged(content) {
		return probeResult{}, fmt.Errorf("status %d: %w", status, errChallenged)
	}
	if status != http.StatusOK {
		return probeResult{}, fmt.Errorf("unexpected status %d: %.120q", status, content)
	}
	return probeResult{status: status, body: content}, nil
}

func findCourseID(fetch fetchFunc, portal, courseSlug string) (int, error) {
	result, err := fetch(fmt.Sprintf(courseSearchURLTemplate, portal, url.QueryEscape(courseSlug)))
	if err != nil {
		return 0, fmt.Errorf("search course %q: %w", courseSlug, err)
	}
	var search courseSearchResponse
	if err := json.Unmarshal(result.body, &search); err != nil {
		return 0, fmt.Errorf("decode course search: %w", err)
	}
	for _, course := range search.Results {
		if course.PublishedTitle == courseSlug {
			return course.ID, nil
		}
	}
	return 0, fmt.Errorf("course %q not in %d subscribed results", courseSlug, len(search.Results))
}

func countCurriculumItems(body []byte) (int, error) {
	var curriculum curriculumResponse
	if err := json.Unmarshal(body, &curriculum); err != nil {
		return 0, fmt.Errorf("decode curriculum: %w", err)
	}
	return len(curriculum.Results), nil
}

func isChallenged(body []byte) bool {
	text := string(body)
	for _, marker := range challengeMarkers {
		if strings.Contains(text, marker) {
			return true
		}
	}
	return false
}

func loadNetscapeCookies(path string, now time.Time) ([]*http.Cookie, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open cookies file: %w", err)
	}
	defer file.Close()

	var cookies []*http.Cookie
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimPrefix(scanner.Text(), httpOnlyPrefix)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != netscapeFieldCount {
			continue
		}
		expiresUnix, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse expiry of cookie %q: %w", fields[5], err)
		}
		if expiresUnix > 0 && expiresUnix < now.Unix() {
			continue
		}
		cookies = append(cookies, &http.Cookie{
			Domain: fields[0],
			Path:   fields[2],
			Secure: fields[3] == "TRUE",
			Name:   fields[5],
			Value:  fields[6],
		})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read cookies file: %w", err)
	}
	return cookies, nil
}
