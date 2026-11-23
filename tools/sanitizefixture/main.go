package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const (
	fixtureFileMode = 0o644
	fixtureDirMode  = 0o750
)

var errPortalRequired = errors.New("-portal is required")

var pythonOracleFiles = []string{"course_content.json", "_udemy.json"}

func main() {
	sourceDir := flag.String("in", "", "Python saved/ directory holding the raw fixtures")
	targetDir := flag.String("out", "testdata/udemy/python", "directory for sanitized fixtures")
	portal := flag.String("portal", "", "real portal subdomain to replace with "+fixturePortal)
	flag.Parse()

	if err := run(*sourceDir, *targetDir, *portal); err != nil {
		fmt.Fprintln(os.Stderr, "sanitizefixture:", err)
		os.Exit(1)
	}
}

func run(sourceDir, targetDir, portal string) error {
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

	for _, name := range pythonOracleFiles {
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
