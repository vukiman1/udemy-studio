package subtitle

import (
	"bytes"
	"fmt"
	"strings"
)

const (
	srtCueFormat       = "%d\n%s --> %s\n%s\n\n"
	srtTimestampFormat = "%02d:%02d:%02d,%03d"
)

// ConvertVTTToSRT returns the SRT bytes the Python downloader writes for a VTT caption it can convert.
func ConvertVTTToSRT(vtt []byte) ([]byte, error) {
	cues, err := parseVTT(vtt)
	if err != nil {
		return nil, err
	}
	var srt bytes.Buffer
	for position, cue := range cues {
		text := make([]string, 0, len(cue.lines))
		for _, line := range cue.lines {
			text = append(text, stripCueTags(line))
		}
		fmt.Fprintf(&srt, srtCueFormat, position+1, cue.start.srt(), cue.end.srt(), unescapeHTML(strings.Join(text, "\n")))
	}
	return srt.Bytes(), nil
}

func (timestamp cueTime) srt() string {
	return fmt.Sprintf(srtTimestampFormat, timestamp.hours, timestamp.minutes, timestamp.seconds, timestamp.milliseconds)
}
