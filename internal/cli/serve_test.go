package cli

import (
	"errors"
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
			OutputDir:     "shadow",
			ChromeDataDir: "/already/absolute",
		},
	}
	if err := resolvePaths(&options); err != nil {
		t.Fatalf("resolvePaths: %v", err)
	}
	want := map[string]string{
		"library": filepath.Join(workingDir, "downloads"),
		"data":    filepath.Join(workingDir, "data"),
		"output":  filepath.Join(workingDir, "shadow"),
		"chrome":  "/already/absolute",
	}
	got := map[string]string{
		"library": options.LibraryDir,
		"data":    options.DataDir,
		"output":  options.Downloads.OutputDir,
		"chrome":  options.Downloads.ChromeDataDir,
	}
	for name, path := range want {
		if got[name] != path {
			t.Errorf("%s = %q, want %q", name, got[name], path)
		}
	}
}

func TestPublicHostsAreBareLowercaseHostnames(t *testing.T) {
	hosts, err := publicHosts([]string{" Studio.Example.com ", ""})
	if err != nil || len(hosts) != 1 || hosts[0] != "studio.example.com" {
		t.Errorf("publicHosts = %v, %v; want [studio.example.com]", hosts, err)
	}
	for _, value := range []string{"https://studio.example.com", "studio.example.com:443", "studio.example.com/app"} {
		if _, err := publicHosts([]string{value}); !errors.Is(err, ErrInvalidPublicHost) {
			t.Errorf("publicHosts(%q) = %v; want ErrInvalidPublicHost", value, err)
		}
	}
}
