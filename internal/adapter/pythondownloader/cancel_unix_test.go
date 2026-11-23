//go:build unix

package pythondownloader

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	childPIDFileEnv  = "FAKE_CHILD_PID_FILE"
	pollInterval     = 10 * time.Millisecond
	childExitTimeout = 5 * time.Second
)

func TestCancellingStopsTheWholeProcessGroup(t *testing.T) {
	downloader, config := newDownloader(t, "wait-for-interrupt")
	childPIDFile := filepath.Join(t.TempDir(), "child.pid")
	t.Setenv(childPIDFileEnv, childPIDFile)
	events := &recorder{}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- downloader.Download(ctx, testJob(), secret.New(tokenValue), events.record) }()

	deadline := time.Now().Add(childExitTimeout)
	for len(events.snapshot()) == 0 && time.Now().Before(deadline) {
		time.Sleep(pollInterval)
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, want context.Canceled", err)
		}
	case <-time.After(interruptGrace):
		t.Fatal("download did not stop after cancel")
	}
	log, _ := os.ReadFile(filepath.Join(config.LogDir, "job1.log"))
	if !strings.Contains(string(log), "interrupted") {
		t.Errorf("python did not receive SIGINT; log:\n%s", log)
	}
	if pid := readChildPID(t, childPIDFile); !exitsWithin(pid, childExitTimeout) {
		t.Errorf("child process %d that ignores SIGINT is still running", pid)
	}
}

func readChildPID(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path) //nolint:gosec // reason: test temp dir
	if err != nil {
		t.Fatalf("read child pid: %v", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		t.Fatalf("parse child pid %q: %v", raw, err)
	}
	return pid
}

func exitsWithin(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if errors.Is(syscall.Kill(pid, 0), syscall.ESRCH) {
			return true
		}
		time.Sleep(pollInterval)
	}
	return false
}
