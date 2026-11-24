package translate

import (
	"fmt"

	"github.com/vukiman1/udemy-studio/internal/subtitle"
)

// CueLines is what a translator outside the app receives for a lecture: one "<index>\t<text>" line per cue.
func CueLines(cues []subtitle.Cue) string {
	return cueLines(cues)
}

// Accept checks a translation made outside the app as strictly as the app checks its own,
// and returns the cues with Vietnamese text and the original timings.
func Accept(cues []subtitle.Cue, translation string) ([]subtitle.Cue, error) {
	translated, err := parseReply(translation, cues)
	if err != nil {
		return nil, err
	}
	result := make([]subtitle.Cue, len(cues))
	for position, cue := range cues {
		cue.Text = translated[cue.Index]
		result[position] = cue
	}
	if spans := driftSpans(texts(cues), texts(result)); len(spans) > 0 {
		return nil, fmt.Errorf("%w: cues %d-%d", ErrDrift, cues[spans[0].from].Index, cues[spans[0].to-1].Index)
	}
	return result, nil
}
