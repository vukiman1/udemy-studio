package subtitle

import "time"

type Cue struct {
	Index int
	Start time.Duration
	End   time.Duration
	Text  string
}
