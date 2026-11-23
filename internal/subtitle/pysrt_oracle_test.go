package subtitle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	libraryDirEnv   = "LIBRARY_DIR"
	pysrtGoldenPath = "../../testdata/subtitle/pysrt_golden.json"
)

type pysrtDigest struct {
	CueCount     int    `json:"cue_count"`
	TimingSHA256 string `json:"timing_sha256"`
	TextSHA256   string `json:"text_sha256"`
}

func TestParseSRTMatchesPysrtOnRealLibrary(t *testing.T) {
	libraryDir := os.Getenv(libraryDirEnv)
	if libraryDir == "" {
		t.Skipf("set %s to the downloads/ directory to compare against pysrt", libraryDirEnv)
	}
	raw, err := os.ReadFile(pysrtGoldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var golden map[string]pysrtDigest
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("decode golden: %v", err)
	}
	library, err := os.OpenRoot(libraryDir)
	if err != nil {
		t.Fatalf("open %s: %v", libraryDir, err)
	}
	defer library.Close()

	for relativePath, want := range golden {
		t.Run(filepath.Base(relativePath), func(t *testing.T) {
			file, err := library.Open(relativePath)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			defer file.Close()
			cues, err := ParseSRT(file)
			if err != nil {
				t.Fatalf("parse: %v", err)
			}
			if got := digestCues(cues); got != want {
				t.Errorf("digest = %+v, want %+v", got, want)
			}
		})
	}
}

func digestCues(cues []Cue) pysrtDigest {
	timings := make([]string, 0, len(cues))
	texts := make([]string, 0, len(cues))
	for _, cue := range cues {
		timings = append(timings, fmt.Sprintf("%d\t%d\t%d", cue.Index, cue.Start.Milliseconds(), cue.End.Milliseconds()))
		texts = append(texts, cue.Text)
	}
	return pysrtDigest{CueCount: len(cues), TimingSHA256: sha256Lines(timings), TextSHA256: sha256Lines(texts)}
}

func sha256Lines(lines []string) string {
	sum := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}
