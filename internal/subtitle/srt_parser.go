package subtitle

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const byteOrderMark = "\ufeff"

var ErrMalformedSRT = errors.New("malformed srt")

var timingLine = regexp.MustCompile(`^(\d{2,}):(\d{2}):(\d{2})[,.](\d{3})\s*-->\s*(\d{2,}):(\d{2}):(\d{2})[,.](\d{3})`)

type ParseError struct {
	Line   int
	Reason string
}

func (parseErr *ParseError) Error() string {
	return fmt.Sprintf("%s: line %d: %s", ErrMalformedSRT, parseErr.Line, parseErr.Reason)
}

func (parseErr *ParseError) Unwrap() error { return ErrMalformedSRT }

type numberedLine struct {
	number int
	text   string
}

func ParseSRTFile(files fs.FS, path string) ([]Cue, error) {
	file, err := files.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open subtitle: %w", err)
	}
	defer file.Close()
	return ParseSRT(file)
}

func ParseSRT(reader io.Reader) ([]Cue, error) {
	blocks, err := splitBlocks(reader)
	if err != nil {
		return nil, err
	}
	var cues []Cue
	for _, block := range blocks {
		cue, err := parseBlock(block)
		if err != nil {
			return nil, err
		}
		cues = append(cues, cue)
	}
	return cues, nil
}

func splitBlocks(reader io.Reader) ([][]numberedLine, error) {
	var blocks [][]numberedLine
	var current []numberedLine
	scanner := bufio.NewScanner(reader)
	for number := 1; scanner.Scan(); number++ {
		text := strings.TrimRight(scanner.Text(), "\r")
		if number == 1 {
			text = strings.TrimPrefix(text, byteOrderMark)
		}
		if strings.TrimSpace(text) == "" {
			if current != nil {
				blocks = append(blocks, current)
				current = nil
			}
			continue
		}
		current = append(current, numberedLine{number: number, text: text})
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read srt: %w", err)
	}
	if current != nil {
		blocks = append(blocks, current)
	}
	return blocks, nil
}

func parseBlock(block []numberedLine) (Cue, error) {
	header := block[0]
	index, err := strconv.Atoi(strings.TrimSpace(header.text))
	if err != nil {
		return Cue{}, &ParseError{Line: header.number, Reason: fmt.Sprintf("cue index %q is not a number", header.text)}
	}
	if len(block) < 2 {
		return Cue{}, &ParseError{Line: header.number, Reason: fmt.Sprintf("cue %d has no timing line", index)}
	}
	timing := block[1]
	start, end, err := parseTiming(timing.text)
	if err != nil {
		return Cue{}, &ParseError{Line: timing.number, Reason: err.Error()}
	}
	texts := make([]string, 0, len(block)-2)
	for _, line := range block[2:] {
		texts = append(texts, line.text)
	}
	return Cue{Index: index, Start: start, End: end, Text: strings.Join(texts, "\n")}, nil
}

func parseTiming(text string) (time.Duration, time.Duration, error) {
	match := timingLine.FindStringSubmatch(strings.TrimSpace(text))
	if match == nil {
		return 0, 0, fmt.Errorf("timing %q is not HH:MM:SS,mmm --> HH:MM:SS,mmm", text)
	}
	return timestamp(match[1:5]), timestamp(match[5:9]), nil
}

func timestamp(parts []string) time.Duration {
	units := [...]time.Duration{time.Hour, time.Minute, time.Second, time.Millisecond}
	var total time.Duration
	for position, part := range parts {
		value, _ := strconv.Atoi(part)
		total += time.Duration(value) * units[position]
	}
	return total
}
