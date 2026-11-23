package subtitle

import (
	"cmp"
	"slices"
	"time"
)

type Track struct {
	Language string
	Cues     []Cue
}

type Line struct {
	Language string
	Text     string
}

type MergedCue struct {
	Start time.Duration
	End   time.Duration
	Lines []Line
}

type cueSpan struct {
	start time.Duration
	end   time.Duration
}

func Merge(tracks []Track) []MergedCue {
	positions := map[cueSpan]int{}
	var merged []MergedCue
	for _, track := range tracks {
		for _, cue := range track.Cues {
			line := Line{Language: track.Language, Text: cue.Text}
			span := cueSpan{start: cue.Start, end: cue.End}
			if position, ok := positions[span]; ok {
				merged[position].Lines = append(merged[position].Lines, line)
				continue
			}
			positions[span] = len(merged)
			merged = append(merged, MergedCue{Start: cue.Start, End: cue.End, Lines: []Line{line}})
		}
	}
	slices.SortStableFunc(merged, func(left, right MergedCue) int {
		return cmp.Or(cmp.Compare(left.Start, right.Start), cmp.Compare(left.End, right.End))
	})
	return merged
}
