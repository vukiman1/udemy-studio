package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

const (
	fixtureFileMode = 0o644
	fixtureDirMode  = 0o750
)

var errPortalRequired = errors.New("-portal is required")

var (
	pythonOracleFiles = []string{"course_content.json", "_udemy.json"}
	hlsMasterPlaylist = regexp.MustCompile(`^index_\d+\.m3u8$`)
)

func main() {
	sourceDir := flag.String("in", "", "Python saved/ directory holding the raw fixtures")
	targetDir := flag.String("out", "testdata/udemy/python", "directory for sanitized fixtures")
	hlsSourceDir := flag.String("hls-in", "", "Python temp/ directory holding the HLS master playlists")
	hlsTargetDir := flag.String("hls-out", "testdata/udemy/hls", "directory for sanitized HLS master playlists")
	portal := flag.String("portal", "", "real portal subdomain to replace with "+fixturePortal)
	flag.Parse()

	err := run(*sourceDir, *targetDir, *portal, pythonOracleFiles)
	if err == nil && *hlsSourceDir != "" {
		err = runHLS(*hlsSourceDir, *hlsTargetDir, *portal)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "sanitizefixture:", err)
		os.Exit(1)
	}
}

func runHLS(sourceDir, targetDir, portal string) error {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("list hls dir: %w", err)
	}
	var masters []string
	for _, entry := range entries {
		if hlsMasterPlaylist.MatchString(entry.Name()) {
			masters = append(masters, entry.Name())
		}
	}
	return run(sourceDir, targetDir, portal, masters)
}

func run(sourceDir, targetDir, portal string, names []string) error {
	if portal == "" {
		return errPortalRequired
	}
	if err := os.MkdirAll(targetDir, fixtureDirMode); err != nil {
		return fmt.Errorf("create %s: %w", targetDir, err)
	}
	source, err := os.OpenRoot(sourceDir)
	if err != nil {
		return fmt.Errorf("open source dir: %w", err)
	}
	defer source.Close()
	target, err := os.OpenRoot(targetDir)
	if err != nil {
		return fmt.Errorf("open target dir: %w", err)
	}
	defer target.Close()

	for _, name := range names {
		raw, err := source.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read fixture %s: %w", name, err)
		}
		if err := target.WriteFile(name, []byte(sanitize(string(raw), portal)), fixtureFileMode); err != nil {
			return fmt.Errorf("write fixture %s: %w", filepath.Join(targetDir, name), err)
		}
		fmt.Println("wrote", filepath.Join(targetDir, name))
	}
	return nil
}
