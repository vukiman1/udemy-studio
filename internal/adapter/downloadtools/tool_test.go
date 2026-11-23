package downloadtools

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func fakeBinary(t *testing.T) string {
	t.Helper()
	binary, err := filepath.Abs("testdata/print-args.sh")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}
	return binary
}

func printedArguments(output *strings.Builder) []string {
	return strings.Split(strings.TrimSuffix(output.String(), "\n"), "\n")
}

func TestToolsGetThePythonDownloadersArguments(t *testing.T) {
	var aria2cOutput, ytDLPOutput strings.Builder
	if err := (Aria2c{Binary: fakeBinary(t), Output: &aria2cOutput}).FetchFile(t.Context(), "https://cdn.example/a.mp4", "/out/ch", "001 A.mp4"); err != nil {
		t.Fatalf("FetchFile: %v", err)
	}
	wantAria2c := []string{
		"https://cdn.example/a.mp4", "-o", "001 A.mp4", "-d", "/out/ch",
		"-j16", "-s20", "-x16", "-c", "--auto-file-renaming=false", "--summary-interval=0", "--disable-ipv6", "--follow-torrent=false",
	}
	if got := printedArguments(&aria2cOutput); !slices.Equal(got, wantAria2c) {
		t.Errorf("aria2c arguments = %q", got)
	}
	if err := (YTDLP{Binary: fakeBinary(t), Output: &ytDLPOutput}).FetchHLS(t.Context(), "file:///tmp/index_1_1280x720.m3u8", "/out/ch/001 A.%(ext)s"); err != nil {
		t.Fatalf("FetchHLS: %v", err)
	}
	wantYTDLP := []string{
		"--enable-file-urls", "--force-generic-extractor", "--concurrent-fragments", "1",
		"--downloader", "aria2c", "--downloader-args", `aria2c:"--disable-ipv6"`,
		"-o", "/out/ch/001 A.%(ext)s", "file:///tmp/index_1_1280x720.m3u8",
	}
	if got := printedArguments(&ytDLPOutput); !slices.Equal(got, wantYTDLP) {
		t.Errorf("yt-dlp arguments = %q", got)
	}
}

func TestToolExitCodeIsAnError(t *testing.T) {
	var output strings.Builder
	if err := (Aria2c{Binary: fakeBinary(t), Output: &output}).FetchFile(t.Context(), "https://cdn.example/fails", "/out", "x"); err == nil {
		t.Error("a non-zero exit was not reported")
	}
}
