package app

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

const testGracePeriod = 200 * time.Millisecond

func TestServeStopsEvenWhileAVideoIsStillStreaming(t *testing.T) {
	listener, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	streaming := make(chan struct{})
	endlessStream := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		writer.(http.Flusher).Flush()
		close(streaming)
		<-request.Context().Done()
	})

	ctx, stop := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() { served <- serve(ctx, listener, endlessStream, slog.New(slog.DiscardHandler), testGracePeriod) }()

	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+listener.Addr().String()+"/media/x", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("start stream: %v", err)
	}
	defer response.Body.Close()
	<-streaming

	stop()
	select {
	case err := <-served:
		if err != nil {
			t.Errorf("serve returned %v, want a clean stop", err)
		}
	case <-time.After(10 * testGracePeriod):
		t.Fatal("serve did not stop while a stream was open")
	}
}
