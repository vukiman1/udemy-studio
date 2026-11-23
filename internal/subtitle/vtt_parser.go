package subtitle

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

const (
	vttSignature      = "WEBVTT"
	cueArrow          = "-->"
	millisecondDigits = 3
	maxMinuteOrSecond = 59
	decimalBase       = 10
)

var ErrNotWebVTT = errors.New("not a WebVTT file: the first line does not start with WEBVTT")

// The character classes are Python's \s and \d, which webvtt-py's patterns use on str.
const (
	pythonSpace     = `[\t\n\v\f\r\x1c-\x1f\x85\p{Z}]`
	cueTimestamp    = `(?:\p{Nd}+:)?\p{Nd}{1,2}:\p{Nd}{1,2}\.\p{Nd}{1,3}`
	cueTimingsRegex = `^` + pythonSpace + `*(` + cueTimestamp + `)` + pythonSpace + `*` + cueArrow + pythonSpace + `*(` + cueTimestamp + `)`
)

var (
	lineTerminator = regexp.MustCompile(`\r\n|\r|\n`)
	cueTimings     = regexp.MustCompile(cueTimingsRegex)
	timestampParts = regexp.MustCompile(`^(?:(\p{Nd}{1,10}):)?(\p{Nd}{1,2}):(\p{Nd}{1,2})\.(\p{Nd}{1,3})$`)
)

type cueTime struct {
	hours, minutes, seconds, milliseconds int64
}

type vttCue struct {
	start, end cueTime
	lines      []string
}

// parseVTT skips malformed cue blocks instead of failing, as webvtt-py 0.6.0 does when not strict.
func parseVTT(data []byte) ([]vttCue, error) {
	text, err := decodeText(data)
	if err != nil {
		return nil, err
	}
	lines := splitLines(text)
	if len(lines) > 0 {
		lines[0] = strings.TrimPrefix(lines[0], byteOrderMark)
	}
	if len(lines) == 0 || !strings.HasPrefix(lines[0], vttSignature) {
		return nil, ErrNotWebVTT
	}
	var cues []vttCue
	for _, block := range vttBlocks(lines) {
		if cue, ok := parseCueBlock(block); ok {
			cues = append(cues, cue)
		}
	}
	return cues, nil
}

func splitLines(text string) []string {
	lines := lineTerminator.Split(text, -1)
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// vttBlocks splits at empty lines and at a "-->" line that cannot be the block's timings (webvtt-py's header split gives the same cues).
func vttBlocks(lines []string) [][]string {
	var blocks [][]string
	var block []string
	for _, line := range lines {
		if (line == "" || startsNewBlock(block, line)) && len(block) > 0 {
			blocks = append(blocks, block)
			block = nil
		}
		if line != "" {
			block = append(block, line)
		}
	}
	if len(block) > 0 {
		blocks = append(blocks, block)
	}
	return blocks
}

func startsNewBlock(block []string, line string) bool {
	return len(block) > 0 && strings.Contains(line, cueArrow) && (len(block) > 1 || strings.Contains(block[0], cueArrow))
}

func isCueBlock(lines []string) bool {
	hasArrow := func(position int) bool { return position < len(lines) && strings.Contains(lines[position], cueArrow) }
	timingsFirst := len(lines) >= 1 && cueTimings.MatchString(lines[0]) && !hasArrow(1)
	timingsAfterIdentifier := len(lines) >= 2 && !hasArrow(0) && cueTimings.MatchString(lines[1]) && !hasArrow(2)
	return timingsFirst || timingsAfterIdentifier
}

func parseCueBlock(lines []string) (vttCue, bool) {
	if !isCueBlock(lines) {
		return vttCue{}, false
	}
	var startText, endText string
	var payload []string
	for _, line := range lines {
		if timings := cueTimings.FindStringSubmatch(line); timings != nil {
			startText, endText = timings[1], timings[2]
			continue
		}
		if startText != "" {
			payload = append(payload, line)
		}
	}
	start, startOK := parseTimestamp(startText)
	end, endOK := parseTimestamp(endText)
	return vttCue{start: start, end: end, lines: payload}, startOK && endOK
}

func parseTimestamp(text string) (cueTime, bool) {
	parts := timestampParts.FindStringSubmatch(text)
	if parts == nil {
		return cueTime{}, false
	}
	fraction := parts[4]
	timestamp := cueTime{
		hours:        decimalValue(parts[1]),
		minutes:      decimalValue(parts[2]),
		seconds:      decimalValue(parts[3]),
		milliseconds: decimalValue(fraction + strings.Repeat("0", millisecondDigits-len([]rune(fraction)))),
	}
	return timestamp, timestamp.minutes <= maxMinuteOrSecond && timestamp.seconds <= maxMinuteOrSecond
}

func decimalValue(digits string) int64 {
	var value int64
	for _, digit := range digits {
		value = value*decimalBase + int64(digitValue(digit))
	}
	return value
}

// Unicode keeps every set of decimal digits as a contiguous run from zero to nine.
func digitValue(digit rune) rune {
	runStart := digit
	for unicode.Is(unicode.Nd, runStart-1) {
		runStart--
	}
	return (digit - runStart) % decimalBase
}

func stripCueTags(line string) string {
	var text strings.Builder
	position := 0
	for {
		start := strings.IndexByte(line[position:], '<')
		if start < 0 {
			break
		}
		start += position
		end := strings.IndexByte(line[start+1:], '>')
		if end < 0 {
			break
		}
		text.WriteString(line[position:start])
		position = start + 1 + end + 1
	}
	text.WriteString(line[position:])
	return text.String()
}
