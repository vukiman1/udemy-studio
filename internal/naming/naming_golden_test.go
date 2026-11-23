package naming

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"
)

const namingGoldenPath = "../../testdata/naming/naming_golden.json"

type namingCase struct {
	Input       string `json:"input"`
	Sanitized   string `json:"sanitized"`
	Deemojified string `json:"deemojified"`
}

func loadNamingCases(t *testing.T) []namingCase {
	t.Helper()
	raw, err := os.ReadFile(namingGoldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	var cases []namingCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("decode golden: %v", err)
	}
	return cases
}

func TestSanitizeFilenameMatchesPathvalidate(t *testing.T) {
	for _, testCase := range loadNamingCases(t) {
		if got := SanitizeFilename(testCase.Input); got != testCase.Sanitized {
			t.Errorf("SanitizeFilename(%s) = %s, pathvalidate gives %s", strconv.Quote(testCase.Input), strconv.Quote(got), strconv.Quote(testCase.Sanitized))
		}
	}
}

func TestRemoveEmojiMatchesDemoji(t *testing.T) {
	for _, testCase := range loadNamingCases(t) {
		if got := RemoveEmoji(testCase.Input); got != testCase.Deemojified {
			t.Errorf("RemoveEmoji(%s) = %s, demoji gives %s", strconv.Quote(testCase.Input), strconv.Quote(got), strconv.Quote(testCase.Deemojified))
		}
	}
}
