// Package transcode re-encodes a course's H.264 videos to H.265 in place, keeping an original whenever the result is not clearly better.
package transcode

import (
	"context"
	"errors"
	"io/fs"
	"time"
)

const (
	H264Codec = "h264"
	HEVCCodec = "hevc"
)

var (
	ErrCourseNotFound = errors.New("course not found")
	ErrAlreadyQueued  = errors.New("course is already being compressed")
	ErrNotQueued      = errors.New("course is not being compressed")
)

type Media struct {
	Codec     string
	Duration  time.Duration
	SizeBytes int64
}

type Prober interface {
	// Probe reads a library video, by its path relative to the library.
	Probe(ctx context.Context, path string) (Media, error)
}

type Encoder interface {
	// Encode writes source as H.265 to target, both relative to the library, reporting how much is encoded.
	Encode(ctx context.Context, source, target string, onProgress func(encoded time.Duration)) error
}

// Files changes the library; *os.Root satisfies it.
type Files interface {
	Stat(path string) (fs.FileInfo, error)
	Rename(oldPath, newPath string) error
	Remove(path string) error
}

// Summary is what is left to compress in a course, from the cache alone; Unchecked videos are not read yet.
type Summary struct {
	Videos    int
	Pending   int
	Unchecked int
}

// Skip remembers a video whose H.265 encode was not worth keeping, until the file changes size.
type Skip struct {
	Path      string
	SizeBytes int64
	Reason    string
}

type SkipStore interface {
	Find(ctx context.Context, path string) (Skip, bool, error)
	Save(ctx context.Context, skip Skip) error
}

type Candidate struct {
	LectureTitle string
	Path         string
	Media        Media
}

type Plan struct {
	Candidates  []Candidate
	AlreadyHEVC int
	// OtherCodec counts videos in neither H.264 nor H.265, left alone rather than re-encoded from another lossy format.
	OtherCodec int
	NotWorthIt int
	// Unreadable counts videos ffprobe cannot read, such as a download left unfinished.
	Unreadable int
}

type State string

const (
	StateQueued    State = "queued"
	StateRunning   State = "running"
	StateDone      State = "done"
	StateCancelled State = "cancelled"
)

func (state State) IsActive() bool { return state == StateQueued || state == StateRunning }

type Current struct {
	LectureTitle string
	Encoded      time.Duration
	Length       time.Duration
}

type Outcome struct {
	Path         string
	LectureTitle string
	Reason       string
}

type Run struct {
	CourseID    string
	State       State
	Total       int
	Finished    int
	Current     *Current
	BytesBefore int64
	BytesAfter  int64
	Skipped     []Outcome
	Failed      []Outcome
}
