package terminalqr

import (
	"fmt"
	"strings"

	"rsc.io/qr"
)

const quietZoneModules = 2

func Render(text string) (string, error) {
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		return "", fmt.Errorf("encode qr for %d bytes: %w", len(text), err)
	}
	isDark := func(x, y int) bool {
		x, y = x-quietZoneModules, y-quietZoneModules
		return x >= 0 && y >= 0 && x < code.Size && y < code.Size && code.Black(x, y)
	}
	side := code.Size + 2*quietZoneModules
	var output strings.Builder
	for y := 0; y < side; y += 2 {
		for x := range side {
			output.WriteString(halfBlock(isDark(x, y), isDark(x, y+1)))
		}
		output.WriteByte('\n')
	}
	return output.String(), nil
}

func halfBlock(top, bottom bool) string {
	switch {
	case top && bottom:
		return " "
	case top:
		return "▄"
	case bottom:
		return "▀"
	default:
		return "█"
	}
}
