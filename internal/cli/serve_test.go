package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vukiman1/udemy-studio/internal/app"
)

func TestRelativePathsAreResolvedAgainstTheWorkingDirectory(t *testing.T) {
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	options := app.Options{
		LibraryDir: "downloads",
		DataDir:    "data",
		Downloads: app.DownloadOptions{
			Python:        ".venv/bin/python",
			PythonRepo:    "../udemy-downloader",
			OutputDir:     "shadow",
			ChromeDataDir: "/already/absolute",
		},
	}
	if err := resolvePaths(&options); err != nil {
		t.Fatalf("resolvePaths: %v", err)
	}
	want := map[string]string{
		"library":     filepath.Join(workingDir, "downloads"),
		"data":        filepath.Join(workingDir, "data"),
		"python":      filepath.Join(workingDir, ".venv/bin/python"),
		"python repo": filepath.Join(filepath.Dir(workingDir), "udemy-downloader"),
		"output":      filepath.Join(workingDir, "shadow"),
		"chrome":      "/already/absolute",
	}
	got := map[string]string{
		"library":     options.LibraryDir,
		"data":        options.DataDir,
		"python":      options.Downloads.Python,
		"python repo": options.Downloads.PythonRepo,
		"output":      options.Downloads.OutputDir,
		"chrome":      options.Downloads.ChromeDataDir,
	}
	for name, path := range want {
		if got[name] != path {
			t.Errorf("%s = %q, want %q", name, got[name], path)
		}
	}
}

func TestUnsetPathsStayUnset(t *testing.T) {
	options := app.Options{}
	if err := resolvePaths(&options); err != nil {
		t.Fatalf("resolvePaths: %v", err)
	}
	if options.Downloads.Python != "" || options.LibraryDir != "" {
		t.Errorf("unset paths were filled in: %+v", options)
	}
}
