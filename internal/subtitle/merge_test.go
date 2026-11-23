package subtitle

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestMergePairsCuesWithIdenticalTiming(t *testing.T) {
	vietnamese := Track{Language: "vi", Cues: []Cue{
		{Index: 1, Start: ms(0), End: ms(1000), Text: "Xin chào"},
		{Index: 2, Start: ms(1000), End: ms(2000), Text: "Tạm biệt"},
	}}
	english := Track{Language: "en", Cues: []Cue{
		{Index: 1, Start: ms(0), End: ms(1000), Text: "Hello"},
		{Index: 2, Start: ms(1000), End: ms(2000), Text: "Goodbye"},
	}}
	want := []MergedCue{
		{Start: ms(0), End: ms(1000), Lines: []Line{{Language: "vi", Text: "Xin chào"}, {Language: "en", Text: "Hello"}}},
		{Start: ms(1000), End: ms(2000), Lines: []Line{{Language: "vi", Text: "Tạm biệt"}, {Language: "en", Text: "Goodbye"}}},
	}
	if diff := cmp.Diff(want, Merge([]Track{vietnamese, english})); diff != "" {
		t.Errorf("merge mismatch (-want +got):\n%s", diff)
	}
}

func TestMergeKeepsCuesThatOnlyOneTrackHas(t *testing.T) {
	vietnamese := Track{Language: "vi", Cues: []Cue{{Start: ms(500), End: ms(900), Text: "chỉ có VI"}}}
	english := Track{Language: "en", Cues: []Cue{
		{Start: ms(0), End: ms(400), Text: "only EN"},
		{Start: ms(500), End: ms(950), Text: "shifted end"},
	}}
	want := []MergedCue{
		{Start: ms(0), End: ms(400), Lines: []Line{{Language: "en", Text: "only EN"}}},
		{Start: ms(500), End: ms(900), Lines: []Line{{Language: "vi", Text: "chỉ có VI"}}},
		{Start: ms(500), End: ms(950), Lines: []Line{{Language: "en", Text: "shifted end"}}},
	}
	if diff := cmp.Diff(want, Merge([]Track{vietnamese, english})); diff != "" {
		t.Errorf("merge mismatch (-want +got):\n%s", diff)
	}
}
