package translate

import (
	"slices"
	"strings"
	"testing"
)

func sentences(count int) []string {
	texts := make([]string, count)
	for position := range texts {
		texts[position] = strings.Repeat("word ", 1+position*7%11)
	}
	return texts
}

func TestASpanWhoseLinesNoLongerFollowTheOriginalIsFlaggedEvenWithoutAShift(t *testing.T) {
	original := sentences(90)
	scrambled := slices.Clone(original)
	slices.Reverse(scrambled[30:60])
	if spans := driftSpans(original, scrambled); len(spans) == 0 {
		t.Error("a span with merged and reordered lines passed the drift check")
	}
}

func TestCaptionsOfEvenLengthAreNotFlaggedForLengthsThatCannotCorrelate(t *testing.T) {
	original := make([]string, 60)
	translated := make([]string, 60)
	for position := range original {
		original[position] = strings.Repeat("x", 60+position%3)
		translated[position] = strings.Repeat("y", 70-position%5)
	}
	if spans := driftSpans(original, translated); len(spans) != 0 {
		t.Errorf("fixed-width captions flagged: %+v", spans)
	}
}

func TestTheLastCuesOfALectureAreCheckedForDrift(t *testing.T) {
	original := sentences(125)
	shiftedTail := slices.Clone(original)
	// Cues 108-124 slip by one: the windows stepping by 10 end at cue 120 and see too little of it.
	for position := 108; position < 124; position++ {
		shiftedTail[position] = original[position+1]
	}
	if spans := driftSpans(original, shiftedTail); len(spans) != 1 || spans[0].to != len(original) {
		t.Errorf("spans = %+v; want only the tail window, ending at cue %d, flagged", spans, len(original))
	}
}
