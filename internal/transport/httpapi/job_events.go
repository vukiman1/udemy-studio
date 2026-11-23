package httpapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/vukiman1/udemy-studio/internal/jobs"
)

const (
	eventStreamContentType = "text/event-stream"
	jobEventName           = "job"
	keepAliveInterval      = 25 * time.Second
)

type jobEventStream struct {
	jobs *jobs.Manager
}

func (stream jobEventStream) serve(writer http.ResponseWriter, request *http.Request) {
	if stream.jobs == nil {
		writeProblem(writer, http.StatusServiceUnavailable, downloadsUnavailableDetail)
		return
	}
	flusher, ok := writer.(http.Flusher)
	if !ok {
		writeProblem(writer, http.StatusInternalServerError, "streaming is not supported by this connection")
		return
	}
	writer.Header().Set("Content-Type", eventStreamContentType)
	writer.Header().Set("Cache-Control", "no-cache")
	updates, unsubscribe := stream.jobs.Subscribe()
	defer unsubscribe()

	for _, job := range stream.jobs.List() {
		if writeJobEvent(writer, job) != nil {
			return
		}
	}
	flusher.Flush()

	keepAlive := time.NewTicker(keepAliveInterval)
	defer keepAlive.Stop()
	for {
		select {
		case <-request.Context().Done():
			return
		case job, isOpen := <-updates:
			if !isOpen || writeJobEvent(writer, job) != nil {
				return
			}
		case <-keepAlive.C:
			if _, err := io.WriteString(writer, ": keep-alive\n\n"); err != nil {
				return
			}
		}
		flusher.Flush()
	}
}

func writeJobEvent(writer io.Writer, job jobs.Job) error {
	payload, err := json.Marshal(toJob(job))
	if err != nil {
		return fmt.Errorf("encode job %s event: %w", job.ID, err)
	}
	_, err = fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", jobEventName, payload)
	return err
}
