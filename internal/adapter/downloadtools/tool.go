// Package downloadtools runs aria2c and yt-dlp with the arguments the Python downloader uses.
package downloadtools

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/processgroup"
)

const interruptGrace = 15 * time.Second

func run(ctx context.Context, output io.Writer, binary string, arguments []string) error {
	command := exec.CommandContext(ctx, binary, arguments...) //nolint:gosec // reason: the binary is operator configuration and arguments come from the download plan
	command.Stdout, command.Stderr = output, output
	processgroup.Interruptible(command, interruptGrace)
	if err := command.Start(); err != nil {
		return fmt.Errorf("start %s: %w", filepath.Base(binary), err)
	}
	err := command.Wait()
	processgroup.Kill(command.Process)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("%s failed: %w", filepath.Base(binary), err)
	}
	return nil
}
