package library

import (
	"strings"
	"unicode"
)

const (
	zeroWidthJoiner     = '‍'
	firstVariationMark  = '︀'
	lastVariationMark   = '️'
	firstSkinToneMarker = '\U0001f3fb'
	lastSkinToneMarker  = '\U0001f3ff'
)

// matchKey compares names across files where Python stripped emoji from the video name but not from its subtitles.
func matchKey(name string) string {
	withoutEmoji := strings.Map(func(r rune) rune {
		if isEmojiPart(r) {
			return -1
		}
		return r
	}, name)
	return strings.Join(strings.Fields(withoutEmoji), " ")
}

func isEmojiPart(r rune) bool {
	return unicode.Is(unicode.So, r) ||
		r == zeroWidthJoiner ||
		(r >= firstVariationMark && r <= lastVariationMark) ||
		(r >= firstSkinToneMarker && r <= lastSkinToneMarker)
}
