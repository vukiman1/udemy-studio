package app

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

const handoffChapter = "course/01 - Intro"

func handoffCues() []subtitle.Cue {
	cues := make([]subtitle.Cue, 60)
	for position := range cues {
		start := time.Duration(position) * 2 * time.Second
		cues[position] = subtitle.Cue{Index: position + 1, Start: start, End: start + time.Second, Text: strings.Repeat("word ", 1+position*7%11)}
	}
	return cues
}

func writeHandoffLecture(t *testing.T, library, name string, languages ...string) {
	t.Helper()
	var document bytes.Buffer
	if err := subtitle.WriteSRT(&document, handoffCues()); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(library, filepath.FromSlash(handoffChapter))
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name+".mp4"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	for _, language := range languages {
		if err := os.WriteFile(filepath.Join(dir, name+"_"+language+".srt"), document.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// translateLines answers an exported file the way a translator would, shifting every text by shift lines.
func translateLines(t *testing.T, workDir, file string, shift int) {
	t.Helper()
	exported, err := os.ReadFile(filepath.Join(workDir, file+sourceTextSuffix)) //nolint:gosec // reason: a path inside the test's temp dir
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(exported)), "\n")
	var reply strings.Builder
	for position, line := range lines {
		index, _, _ := strings.Cut(line, "\t")
		_, text, _ := strings.Cut(lines[min(position+shift, len(lines)-1)], "\t")
		reply.WriteString(index + "\tVI " + text + "\n")
	}
	if err := os.WriteFile(filepath.Join(workDir, file+translatedSuffix), []byte(reply.String()), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestHandoffWritesOnlyTranslationsThatPassTheAppsChecks(t *testing.T) {
	library, workDir := t.TempDir(), t.TempDir()
	writeHandoffLecture(t, library, "001 Faithful", "en")
	writeHandoffLecture(t, library, "002 Shifted", "en")
	writeHandoffLecture(t, library, "003 Untouched", "en")
	writeHandoffLecture(t, library, "004 Done", "en", "vi")
	handoff := SubtitleHandoff{LibraryDir: library, WorkDir: workDir}

	count, err := handoff.Export("")
	if err != nil || count != 3 {
		t.Fatalf("Export = %d, %v; want the 3 lectures without Vietnamese subtitles", count, err)
	}
	files := exportedFiles(t, workDir)
	translateLines(t, workDir, files["Faithful"], 0)
	translateLines(t, workDir, files["Shifted"], 1)
	report, err := handoff.Import()
	if err != nil {
		t.Fatalf("Import: %v", err)
	}
	if len(report.Written) != 1 || len(report.Rejected) != 1 || len(report.Waiting) != 1 || !strings.Contains(report.Rejected[0], "Shifted") {
		t.Errorf("report = %+v; want Faithful written, Shifted rejected, Untouched waiting", report)
	}
	written, _ := filepath.Glob(filepath.Join(library, filepath.FromSlash(handoffChapter), "*_vi.srt"))
	names := make([]string, len(written))
	for position, path := range written {
		names[position] = filepath.Base(path)
	}
	slices.Sort(names)
	if !slices.Equal(names, []string{"001 Faithful_vi.srt", "004 Done_vi.srt"}) {
		t.Errorf("Vietnamese subtitles in the library = %v", names)
	}
	if again, _ := handoff.Import(); len(again.Written) != 0 {
		t.Errorf("a second import wrote over existing subtitles: %+v", again)
	}
	if _, err := handoff.Export(""); err != nil {
		t.Fatalf("export again: %v", err)
	}
	if again := exportedFiles(t, workDir); again["Shifted"] != files["Shifted"] || again["Untouched"] != files["Untouched"] || len(again) != 2 {
		t.Errorf("exporting again named files %v, want the remaining lectures under their first names %v", again, files)
	}
}

// exportedFiles maps each exported lecture title to its file name, from the manifest Export wrote.
func exportedFiles(t *testing.T, workDir string) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(workDir, handoffManifestName)) //nolint:gosec // reason: a path inside the test's temp dir
	if err != nil {
		t.Fatal(err)
	}
	var entries []handoffEntry
	if err := json.Unmarshal(raw, &entries); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{}
	for _, entry := range entries {
		files[entry.LectureTitle] = entry.File
	}
	return files
}
