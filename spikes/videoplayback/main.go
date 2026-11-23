package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const (
	maxReportBytes    = 64 << 10
	readHeaderTimeout = 10 * time.Second
)

//go:embed player.html
var playerPage []byte

type sampleFiles map[string]string

type playbackReport struct {
	Sample string            `json:"sample"`
	Steps  []json.RawMessage `json:"steps"`
	Result string            `json:"result"`
}

func main() {
	address := flag.String("addr", "127.0.0.1:8765", "listen address")
	hevcPath := flag.String("hevc", "", "HEVC (hvc1) mp4 file")
	h264Path := flag.String("h264", "", "H.264 (avc1) mp4 file")
	flag.Parse()

	samples := sampleFiles{"hevc": *hevcPath, "h264": *h264Path}
	for name, path := range samples {
		if _, err := os.Stat(path); err != nil {
			fmt.Fprintf(os.Stderr, "sample %s: %v\n", name, err)
			os.Exit(1)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", servePlayer)
	mux.HandleFunc("GET /media/{sample}", samples.serveMedia)
	mux.HandleFunc("POST /report", receiveReport)

	server := &http.Server{Addr: *address, Handler: logRequests(mux), ReadHeaderTimeout: readHeaderTimeout}
	slog.Info("listening", "addr", *address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}

func servePlayer(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("content-type", "text/html; charset=utf-8")
	_, _ = writer.Write(playerPage)
}

func (samples sampleFiles) serveMedia(writer http.ResponseWriter, request *http.Request) {
	path, ok := samples[request.PathValue("sample")]
	if !ok {
		http.NotFound(writer, request)
		return
	}
	file, err := os.Open(path)
	if err != nil {
		http.Error(writer, "open sample", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		http.Error(writer, "stat sample", http.StatusInternalServerError)
		return
	}
	writer.Header().Set("content-type", "video/mp4")
	http.ServeContent(writer, request, "", info.ModTime(), file)
}

func receiveReport(writer http.ResponseWriter, request *http.Request) {
	body, err := io.ReadAll(io.LimitReader(request.Body, maxReportBytes))
	if err != nil {
		http.Error(writer, "read report", http.StatusBadRequest)
		return
	}
	var report playbackReport
	if err := json.Unmarshal(body, &report); err != nil {
		http.Error(writer, "decode report", http.StatusBadRequest)
		return
	}
	slog.Info("REPORT", "sample", report.Sample, "result", report.Result, "ua", request.UserAgent())
	for _, step := range report.Steps {
		slog.Info("  step", "detail", string(step))
	}
	writer.WriteHeader(http.StatusNoContent)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (recorder *statusRecorder) WriteHeader(status int) {
	recorder.status = status
	recorder.ResponseWriter.WriteHeader(status)
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
		next.ServeHTTP(recorder, request)
		if request.URL.Path == "/report" {
			return
		}
		slog.Info("request", "path", request.URL.Path, "range", request.Header.Get("range"),
			"status", recorder.status, "content_range", recorder.Header().Get("content-range"))
	})
}
