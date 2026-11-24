package translate

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

const (
	cueSeparator  = "\t"
	maxListedCues = 10
)

// cueLines writes one "<index>\t<text>" line per cue, so the model sees text only and cannot touch timings.
func cueLines(cues []subtitle.Cue) string {
	var lines strings.Builder
	for _, cue := range cues {
		fmt.Fprintf(&lines, "%d%s%s\n", cue.Index, cueSeparator, singleLine(cue.Text))
	}
	return lines.String()
}

// parseReply returns the translated text of every cue, or ErrMalformedReply when any cue is missing, extra, repeated or empty.
func parseReply(reply string, cues []subtitle.Cue) (map[int]string, error) {
	translated := map[int]string{}
	var problems []string
	for _, line := range strings.Split(strings.TrimSpace(reply), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indexText, text, ok := strings.Cut(strings.TrimSpace(line), cueSeparator)
		index, err := strconv.Atoi(strings.TrimSpace(indexText))
		if !ok || err != nil {
			problems = append(problems, fmt.Sprintf("line %q is not <index><TAB><text>", truncate(line)))
			continue
		}
		if _, seen := translated[index]; seen {
			problems = append(problems, fmt.Sprintf("cue %d translated twice", index))
		}
		translated[index] = strings.TrimSpace(text)
	}
	problems = append(problems, coverageProblems(translated, cues)...)
	if len(problems) > 0 {
		return nil, fmt.Errorf("%w: %s", ErrMalformedReply, strings.Join(problems, "; "))
	}
	return translated, nil
}

func coverageProblems(translated map[int]string, cues []subtitle.Cue) []string {
	expected := map[int]bool{}
	var missing, empty, extra []int
	for _, cue := range cues {
		expected[cue.Index] = true
		text, ok := translated[cue.Index]
		switch {
		case !ok:
			missing = append(missing, cue.Index)
		case text == "":
			empty = append(empty, cue.Index)
		}
	}
	for index := range translated {
		if !expected[index] {
			extra = append(extra, index)
		}
	}
	slices.Sort(extra)
	var problems []string
	for _, group := range []struct {
		label   string
		indices []int
	}{{"missing", missing}, {"empty", empty}, {"not asked for", extra}} {
		if len(group.indices) > 0 {
			problems = append(problems, fmt.Sprintf("cues %s: %v", group.label, group.indices[:min(len(group.indices), maxListedCues)]))
		}
	}
	return problems
}

func singleLine(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func truncate(line string) string {
	const shown = 60
	runes := []rune(line)
	if len(runes) <= shown {
		return line
	}
	return string(runes[:shown]) + "…"
}
