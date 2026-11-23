package subtitle

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"
)

const (
	vttFixtureDir     = "subtitle/vtt"
	vttExtension      = ".vtt"
	srtGoldenSuffix   = ".srt"
	errorGoldenSuffix = ".error"
)

var pythonErrors = map[string]error{
	"MalformedFileError": ErrNotWebVTT,
	"UnicodeDecodeError": errUndecodableText,
}

func TestConvertVTTToSRTMatchesPythonDownloader(t *testing.T) {
	testdata := os.DirFS("../../testdata")
	fixtures, err := fs.Glob(testdata, path.Join(vttFixtureDir, "*"+vttExtension))
	if err != nil || len(fixtures) == 0 {
		t.Fatalf("no VTT fixtures: %v", err)
	}
	for _, fixture := range fixtures {
		stem := strings.TrimSuffix(fixture, vttExtension)
		t.Run(path.Base(stem), func(t *testing.T) {
			vtt, err := fs.ReadFile(testdata, fixture)
			if err != nil {
				t.Fatalf("read fixture: %v", err)
			}
			got, convertErr := ConvertVTTToSRT(vtt)
			want, err := fs.ReadFile(testdata, stem+srtGoldenSuffix)
			if errors.Is(err, fs.ErrNotExist) {
				assertPythonError(t, testdata, stem, convertErr)
				return
			}
			if err != nil {
				t.Fatalf("read golden: %v", err)
			}
			if convertErr != nil || !bytes.Equal(got, want) {
				t.Errorf("ConvertVTTToSRT = %q, %v\nwant %q", got, convertErr, want)
			}
		})
	}
}

func assertPythonError(t *testing.T, testdata fs.FS, stem string, convertErr error) {
	t.Helper()
	recorded, err := fs.ReadFile(testdata, stem+errorGoldenSuffix)
	if err != nil {
		t.Fatalf("no %s or %s golden: %v", srtGoldenSuffix, errorGoldenSuffix, err)
	}
	pythonError := strings.TrimSpace(string(recorded))
	want, known := pythonErrors[pythonError]
	if !known {
		t.Fatalf("no Go error mapped to Python's %s", pythonError)
	}
	if !errors.Is(convertErr, want) {
		t.Errorf("ConvertVTTToSRT error = %v, want %v like Python's %s", convertErr, want, pythonError)
	}
}
