package terminalqr

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRenderDrawsASquareOfHalfBlocks(t *testing.T) {
	rendered, err := Render("http://192.168.1.20:8080/?code=ABCDEFGHJK")
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	rows := strings.Split(strings.TrimSuffix(rendered, "\n"), "\n")
	width := utf8.RuneCountInString(rows[0])
	if wantRows := (width + 1) / 2; len(rows) != wantRows {
		t.Errorf("rows = %d, want %d for width %d", len(rows), wantRows, width)
	}
	for index, row := range rows {
		if utf8.RuneCountInString(row) != width || strings.Trim(row, " ▀▄█") != "" {
			t.Fatalf("row %d = %q is not %d half-block cells", index, row, width)
		}
	}
	if !strings.HasPrefix(rows[0], "████") {
		t.Errorf("first row %q should start with the light quiet zone", rows[0])
	}
}
