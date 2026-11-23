package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

const (
	videoExtension        = ".mp4"
	manifestMode          = 0o644
	manifestDirMode       = 0o750
	manifestSchemaVersion = 1
	durationScale         = 1000
)

var errRootRequired = errors.New("-root is required")

type manifest struct {
	SchemaVersion int     `json:"schema_version"`
	Files         []entry `json:"files"`
}

type entry struct {
	Path            string   `json:"path"`
	SizeBytes       int64    `json:"size_bytes"`
	DurationSeconds *float64 `json:"duration_seconds,omitempty"`
	VideoCodec      string   `json:"video_codec,omitempty"`
	VideoCodecTag   string   `json:"video_codec_tag,omitempty"`
	Width           int      `json:"width,omitempty"`
	Height          int      `json:"height,omitempty"`
}

type ffprobeOutput struct {
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
	Streams []struct {
		CodecName string `json:"codec_name"`
		CodecTag  string `json:"codec_tag_string"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func main() {
	root := flag.String("root", "", "downloads/ directory to describe")
	output := flag.String("out", "testdata/library/manifest.json", "manifest file to write")
	flag.Parse()

	if err := run(context.Background(), *root, *output); err != nil {
		fmt.Fprintln(os.Stderr, "librarymanifest:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, root, output string) error {
	if root == "" {
		return errRootRequired
	}
	var files []entry
	err := filepath.WalkDir(root, func(path string, dirEntry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if dirEntry.IsDir() || strings.HasPrefix(dirEntry.Name(), ".") {
			return nil
		}
		described, err := describe(ctx, root, path, dirEntry)
		if err != nil {
			return err
		}
		files = append(files, described)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk %s: %w", root, err)
	}
	slices.SortFunc(files, func(left, right entry) int { return strings.Compare(left.Path, right.Path) })

	encoded, err := json.MarshalIndent(manifest{SchemaVersion: manifestSchemaVersion, Files: files}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(output), manifestDirMode); err != nil {
		return fmt.Errorf("create manifest dir: %w", err)
	}
	if err := os.WriteFile(output, append(encoded, '\n'), manifestMode); err != nil {
		return fmt.Errorf("write manifest %s: %w", output, err)
	}
	fmt.Printf("wrote %s (%d files)\n", output, len(files))
	return nil
}

func describe(ctx context.Context, root, path string, dirEntry fs.DirEntry) (entry, error) {
	info, err := dirEntry.Info()
	if err != nil {
		return entry{}, fmt.Errorf("stat %s: %w", path, err)
	}
	relative, err := filepath.Rel(root, path)
	if err != nil {
		return entry{}, fmt.Errorf("relativize %s: %w", path, err)
	}
	described := entry{Path: filepath.ToSlash(relative), SizeBytes: info.Size()}
	if filepath.Ext(path) != videoExtension {
		return described, nil
	}
	probe, err := runFFprobe(ctx, path)
	if err != nil {
		return entry{}, err
	}
	duration, err := strconv.ParseFloat(probe.Format.Duration, 64)
	if err != nil {
		return entry{}, fmt.Errorf("parse duration %q of %s: %w", probe.Format.Duration, relative, err)
	}
	rounded := math.Round(duration*durationScale) / durationScale
	described.DurationSeconds = &rounded
	if len(probe.Streams) > 0 {
		stream := probe.Streams[0]
		described.VideoCodec, described.VideoCodecTag = stream.CodecName, stream.CodecTag
		described.Width, described.Height = stream.Width, stream.Height
	}
	return described, nil
}

func runFFprobe(ctx context.Context, path string) (ffprobeOutput, error) {
	//nolint:gosec // reason: the binary is fixed and path is an argument, never parsed by a shell
	command := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-select_streams", "v:0",
		"-show_entries", "format=duration:stream=codec_name,codec_tag_string,width,height", "-of", "json", path)
	raw, err := command.Output()
	if err != nil {
		return ffprobeOutput{}, fmt.Errorf("ffprobe %s: %w", path, err)
	}
	var probe ffprobeOutput
	if err := json.Unmarshal(raw, &probe); err != nil {
		return ffprobeOutput{}, fmt.Errorf("decode ffprobe output for %s: %w", path, err)
	}
	return probe, nil
}
