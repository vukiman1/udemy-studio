package subtitle

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func ms(milliseconds int) time.Duration { return time.Duration(milliseconds) * time.Millisecond }

func TestParseSRTReadsCues(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  []Cue
	}{
		{
			name:  "two cues",
			input: "1\n00:00:00,090 --> 00:00:05,400\nHello there\n\n2\n00:00:05,400 --> 00:00:09,570\nSecond\n",
			want: []Cue{
				{Index: 1, Start: ms(90), End: ms(5400), Text: "Hello there"},
				{Index: 2, Start: ms(5400), End: ms(9570), Text: "Second"},
			},
		},
		{
			name:  "byte order mark, CRLF and no trailing newline",
			input: "\ufeff1\r\n01:02:03,004 --> 01:02:04,005\r\nXin chào",
			want:  []Cue{{Index: 1, Start: time.Hour + 2*time.Minute + 3*time.Second + ms(4), End: time.Hour + 2*time.Minute + 4*time.Second + ms(5), Text: "Xin chào"}},
		},
		{
			name:  "multi-line text and extra blank lines",
			input: "\n\n1\n00:00:01,000 --> 00:00:02,000\nline one\nline two\n\n\n\n2\n00:00:03,000 --> 00:00:04,000\nnext\n\n",
			want: []Cue{
				{Index: 1, Start: ms(1000), End: ms(2000), Text: "line one\nline two"},
				{Index: 2, Start: ms(3000), End: ms(4000), Text: "next"},
			},
		},
		{
			name:  "dot as millisecond separator",
			input: "7\n00:00:01.500 --> 00:00:02.250\nx\n",
			want:  []Cue{{Index: 7, Start: ms(1500), End: ms(2250), Text: "x"}},
		},
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			got, err := ParseSRT(strings.NewReader(testCase.input))
			if err != nil {
				t.Fatalf("ParseSRT: %v", err)
			}
			if diff := cmp.Diff(testCase.want, got); diff != "" {
				t.Errorf("cues mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseSRTRejectsMalformedBlocks(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantLine int
	}{
		{name: "index is not a number", input: "one\n00:00:01,000 --> 00:00:02,000\nx\n", wantLine: 1},
		{name: "missing timing line", input: "1\n\n2\n00:00:01,000 --> 00:00:02,000\nx\n", wantLine: 1},
		{name: "bad timestamp", input: "1\n00:00:01,000 --> 00:0x:02,000\nx\n", wantLine: 2},
		{name: "error points at the later block", input: "1\n00:00:01,000 --> 00:00:02,000\nx\n\n2\nnot a timing\ny\n", wantLine: 6},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := ParseSRT(strings.NewReader(testCase.input))
			var parseErr *ParseError
			if !errors.As(err, &parseErr) {
				t.Fatalf("error = %v, want *ParseError", err)
			}
			if !errors.Is(err, ErrMalformedSRT) {
				t.Errorf("error %v does not match ErrMalformedSRT", err)
			}
			if parseErr.Line != testCase.wantLine {
				t.Errorf("line = %d, want %d (%v)", parseErr.Line, testCase.wantLine, err)
			}
		})
	}
}
