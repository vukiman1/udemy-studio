package subtitle

import (
	"strings"
	"testing"
	"time"
)

func TestWriteVTTStacksLanguagesPerCue(t *testing.T) {
	cues := []MergedCue{
		{Start: ms(90), End: ms(5400), Lines: []Line{{Language: "vi", Text: "Xin chào"}, {Language: "en", Text: "Hello"}}},
		{Start: time.Hour + ms(1), End: time.Hour + ms(2), Lines: []Line{{Language: "en", Text: "a < b && c > d"}}},
	}
	want := "WEBVTT\n\n" +
		"00:00:00.090 --> 00:00:05.400\nXin chào\nHello\n\n" +
		"01:00:00.001 --> 01:00:00.002\na &lt; b &amp;&amp; c &gt; d\n\n"

	var output strings.Builder
	if err := WriteVTT(&output, cues); err != nil {
		t.Fatalf("WriteVTT: %v", err)
	}
	if output.String() != want {
		t.Errorf("vtt =\n%q\nwant\n%q", output.String(), want)
	}
}
