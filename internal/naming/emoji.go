package naming

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
)

//go:embed emoji_codes.json
var emojiCodesJSON []byte

// emoji_codes.json is demoji 1.1.0's code list in the order its regex alternation uses (longest first).
var emojiPattern = sync.OnceValue(func() *regexp.Regexp {
	var codes []string
	if err := json.Unmarshal(emojiCodesJSON, &codes); err != nil {
		panic("naming: embedded emoji_codes.json is invalid: " + err.Error())
	}
	quoted := make([]string, len(codes))
	for index, code := range codes {
		quoted[index] = regexp.QuoteMeta(code)
	}
	return regexp.MustCompile(strings.Join(quoted, "|"))
})

// RemoveEmoji reproduces demoji.replace(text, "") as the Python downloader's deEmojify uses it.
func RemoveEmoji(text string) string {
	return emojiPattern().ReplaceAllLiteralString(text, "")
}
