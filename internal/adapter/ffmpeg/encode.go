package ffmpeg

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/vukiman1/udemy-studio/internal/platform/processgroup"
)

const (
	// The Python downloader's encode, but -preset fast: measured as small as medium, near as sharp, and faster.
	constantRateFactor = "28"
	x265Preset         = "fast"
	// Encoding saturates the CPU for hours, so it yields to whatever the learner is doing.
	NiceBinary     = "nice"
	niceness       = "10"
	interruptGrace = 5 * time.Second
	encodedKey     = "out_time_us="
)

func (tools Tools) Encode(ctx context.Context, source, target string, onProgress func(encoded time.Duration)) error {
	arguments := []string{
		"-n", niceness, tools.FFmpeg,
		"-hide_banner", "-nostdin", "-nostats", "-loglevel", "error", "-y",
		"-i", tools.absolute(source),
		"-c:v", "libx265", "-crf", constantRateFactor, "-preset", x265Preset, "-x265-params", "log-level=error",
		"-tag:v", "hvc1", "-c:a", "copy", "-movflags", "+faststart",
		"-progress", "pipe:1", "-f", "mp4", tools.absolute(target),
	}
	command := exec.CommandContext(ctx, NiceBinary, arguments...) //nolint:gosec // reason: the binary is operator configuration and paths come from the library scan
	var errorOutput bytes.Buffer
	command.Stderr = &errorOutput
	progress, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("pipe ffmpeg progress: %w", err)
	}
	processgroup.Interruptible(command, interruptGrace)
	if err := command.Start(); err != nil {
		return fmt.Errorf("start ffmpeg for %s: %w", source, err)
	}
	reportProgress(progress, onProgress)
	err = command.Wait()
	processgroup.Kill(command.Process)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err != nil {
		return fmt.Errorf("ffmpeg could not encode %s: %w: %s", source, err, strings.TrimSpace(errorOutput.String()))
	}
	return nil
}

func reportProgress(progress io.Reader, onProgress func(time.Duration)) {
	lines := bufio.NewScanner(progress)
	for lines.Scan() {
		value, ok := strings.CutPrefix(lines.Text(), encodedKey)
		if !ok {
			continue
		}
		if microseconds, err := strconv.ParseInt(value, 10, 64); err == nil {
			onProgress(time.Duration(microseconds) * time.Microsecond)
		}
	}
}
