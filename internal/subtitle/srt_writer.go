package subtitle

import (
	"fmt"
	"io"
	"time"
)

func WriteSRT(writer io.Writer, cues []Cue) error {
	for _, cue := range cues {
		if _, err := fmt.Fprintf(writer, srtCueFormat, cue.Index, srtTimestamp(cue.Start), srtTimestamp(cue.End), cue.Text); err != nil {
			return fmt.Errorf("write srt cue %d: %w", cue.Index, err)
		}
	}
	return nil
}

func srtTimestamp(offset time.Duration) string {
	hours := offset / time.Hour
	minutes := offset % time.Hour / time.Minute
	seconds := offset % time.Minute / time.Second
	milliseconds := offset % time.Second / time.Millisecond
	return fmt.Sprintf(srtTimestampFormat, hours, minutes, seconds, milliseconds)
}
