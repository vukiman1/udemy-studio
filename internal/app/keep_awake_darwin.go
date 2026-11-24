//go:build darwin

package app

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"strconv"
)

// caffeinate -i keeps macOS from sleeping while it runs. With -w it also
// exits with this process, even one killed before it could stop caffeinate itself.
const caffeinateBinary = "caffeinate"

// keepAwake holds off sleep through a long download until the returned stop is called.
func keepAwake(ctx context.Context, logger *slog.Logger) (stop func()) {
	caffeinate := exec.CommandContext(ctx, caffeinateBinary, "-i", "-w", strconv.Itoa(os.Getpid())) //nolint:gosec // reason: a fixed binary and this process's own pid
	if err := caffeinate.Start(); err != nil {
		if ctx.Err() == nil {
			logger.Warn("the computer may sleep during this download", "err", err)
		}
		return func() {}
	}
	return func() {
		_ = caffeinate.Process.Kill()
		_ = caffeinate.Wait()
	}
}
