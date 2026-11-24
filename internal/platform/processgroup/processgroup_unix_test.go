//go:build unix

package processgroup

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const (
	grace         = 5 * time.Second
	pollInterval  = 10 * time.Millisecond
	startedWithin = 5 * time.Second
	// The leader's SIGINT trap exits with this, as a shell does after Ctrl-C.
	interruptedExitCode = 130
	// The leader traps SIGINT and exits; its child ignores SIGINT, as yt-dlp's aria2c may. The trap comes before the
	// pid file the test waits for, and both loops end after 30s, so a test that dies early leaves nothing running.
	leaderScript = `trap 'exit 130' INT
sh -c 'trap "" INT; i=0; while [ $i -lt 600 ]; do sleep 0.05; i=$((i+1)); done' &
echo $! > "$1"
i=0; while [ $i -lt 600 ]; do sleep 0.05; i=$((i+1)); done`
)

func TestCancellingInterruptsTheLeaderAndKillLeavesNoChildBehind(t *testing.T) {
	childPIDFile := filepath.Join(t.TempDir(), "child.pid")
	ctx, cancel := context.WithCancel(t.Context())
	command := exec.CommandContext(ctx, "sh", "-c", leaderScript, "leader", childPIDFile) //nolint:gosec // reason: a fixed script; the only variable is the test's temp file
	Interruptible(command, grace)
	if err := command.Start(); err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		Kill(command.Process)
		_ = command.Wait()
	})
	childPID := waitForPID(t, childPIDFile)
	started := time.Now()
	cancel()
	var exit *exec.ExitError
	if err := command.Wait(); !errors.As(err, &exit) || exit.ExitCode() != interruptedExitCode {
		t.Errorf("leader ended with %v; want exit %d from its SIGINT trap, so tools get to clean up", err, interruptedExitCode)
	}
	if waited := time.Since(started); waited >= grace {
		t.Errorf("the leader took %v to stop; want it interrupted, not killed after the %v grace", waited, grace)
	}
	Kill(command.Process)
	if !exitsWithin(childPID, grace) {
		t.Errorf("child %d that ignores SIGINT is still running", childPID)
		_ = syscall.Kill(childPID, syscall.SIGKILL)
	}
}

func waitForPID(t *testing.T, path string) int {
	t.Helper()
	deadline := time.Now().Add(startedWithin)
	for time.Now().Before(deadline) {
		raw, err := os.ReadFile(path) //nolint:gosec // reason: a path inside the test's temp dir
		if pid, parseErr := strconv.Atoi(strings.TrimSpace(string(raw))); err == nil && parseErr == nil {
			return pid
		}
		time.Sleep(pollInterval)
	}
	t.Fatal("the leader never started its child")
	return 0
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
