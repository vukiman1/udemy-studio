package subtitle

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

const vttHeader = "WEBVTT\n\n"

var vttTextEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

func WriteVTT(writer io.Writer, cues []MergedCue) error {
	buffered := bufio.NewWriter(writer)
	if _, err := buffered.WriteString(vttHeader); err != nil {
		return fmt.Errorf("write vtt header: %w", err)
	}
	for _, cue := range cues {
		if _, err := fmt.Fprintf(buffered, "%s --> %s\n", vttTimestamp(cue.Start), vttTimestamp(cue.End)); err != nil {
			return fmt.Errorf("write vtt timing: %w", err)
		}
		for _, line := range cue.Lines {
			if _, err := fmt.Fprintln(buffered, vttTextEscaper.Replace(line.Text)); err != nil {
				return fmt.Errorf("write vtt text: %w", err)
			}
		}
		if err := buffered.WriteByte('\n'); err != nil {
			return fmt.Errorf("write vtt separator: %w", err)
		}
	}
	if err := buffered.Flush(); err != nil {
		return fmt.Errorf("flush vtt: %w", err)
	}
	return nil
}

func vttTimestamp(offset time.Duration) string {
	hours := offset / time.Hour
	minutes := offset % time.Hour / time.Minute
	seconds := offset % time.Minute / time.Second
	milliseconds := offset % time.Second / time.Millisecond
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, seconds, milliseconds)
}
