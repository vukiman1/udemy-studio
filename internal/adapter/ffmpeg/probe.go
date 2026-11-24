// Package ffmpeg probes and re-encodes library videos with the ffprobe and ffmpeg binaries.
package ffmpeg

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/vukiman1/udemy-studio/internal/transcode"
)

type Tools struct {
	FFmpeg     string
	FFprobe    string
	LibraryDir string
}

type probeOutput struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
		Size     string `json:"size"`
	} `json:"format"`
}

func (tools Tools) Probe(ctx context.Context, path string) (transcode.Media, error) {
	output, err := exec.CommandContext(ctx, tools.FFprobe, //nolint:gosec // reason: the binary is operator configuration and the path comes from the library scan
		"-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=codec_name:format=duration,size",
		"-of", "json", tools.absolute(path)).Output()
	if err != nil {
		return transcode.Media{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}
	var probed probeOutput
	if err := json.Unmarshal(output, &probed); err != nil {
		return transcode.Media{}, fmt.Errorf("read ffprobe output for %s: %w", path, err)
	}
	if len(probed.Streams) == 0 {
		return transcode.Media{}, fmt.Errorf("ffprobe %s: no video stream", path)
	}
	seconds, err := strconv.ParseFloat(probed.Format.Duration, 64)
	if err != nil {
		return transcode.Media{}, fmt.Errorf("ffprobe %s: duration %q: %w", path, probed.Format.Duration, err)
	}
	size, err := strconv.ParseInt(probed.Format.Size, 10, 64)
	if err != nil {
		return transcode.Media{}, fmt.Errorf("ffprobe %s: size %q: %w", path, probed.Format.Size, err)
	}
	return transcode.Media{
		Codec:     probed.Streams[0].CodecName,
		Duration:  time.Duration(seconds * float64(time.Second)),
		SizeBytes: size,
	}, nil
}

func (tools Tools) absolute(path string) string {
	return filepath.Join(tools.LibraryDir, filepath.FromSlash(path))
}
