package transcode

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"time"
)

const (
	// A re-encode is kept only when it plays as long as the original, give or take this.
	durationTolerance = 500 * time.Millisecond
	temporarySuffix   = ".h265.tmp"
)

type fileResultKind int

const (
	fileCompressed fileResultKind = iota
	fileNotWorthIt
	fileFailed
	fileCancelled
)

type fileResult struct {
	kind       fileResultKind
	reason     string
	bytesAfter int64
}

func (result fileResult) applyTo(run *Run, candidate Candidate) {
	switch result.kind {
	case fileCancelled:
		return
	case fileCompressed:
		run.BytesBefore += candidate.Media.SizeBytes
		run.BytesAfter += result.bytesAfter
	case fileNotWorthIt:
		run.Skipped = append(run.Skipped, Outcome{Path: candidate.Path, LectureTitle: candidate.LectureTitle, Reason: result.reason})
	case fileFailed:
		run.Failed = append(run.Failed, Outcome{Path: candidate.Path, LectureTitle: candidate.LectureTitle, Reason: result.reason})
	}
	run.Finished++
	run.Current = nil
}

// compressFile encodes next to the original, under a hidden name the library scan ignores, and swaps it in only once verified.
func (service *Service) compressFile(ctx context.Context, candidate Candidate, onProgress func(time.Duration)) fileResult {
	temporary := path.Join(path.Dir(candidate.Path), "."+path.Base(candidate.Path)+temporarySuffix)
	defer service.discard(temporary)
	err := service.deps.Encoder.Encode(ctx, candidate.Path, temporary, onProgress)
	if ctx.Err() != nil {
		return fileResult{kind: fileCancelled}
	}
	if err != nil {
		return fileResult{kind: fileFailed, reason: err.Error()}
	}
	// The encode is done, so a cancel from here on must not turn the check of it into a failure.
	verifyCtx := context.WithoutCancel(ctx)
	encoded, err := service.deps.Prober.Probe(verifyCtx, temporary)
	if err != nil {
		return fileResult{kind: fileFailed, reason: fmt.Sprintf("probe the encoded file: %v", err)}
	}
	if reason, ok := notWorthKeeping(candidate.Media, encoded); ok {
		service.remember(verifyCtx, Skip{Path: candidate.Path, SizeBytes: candidate.Media.SizeBytes, Reason: reason})
		return fileResult{kind: fileNotWorthIt, reason: reason}
	}
	if err := service.deps.Files.Rename(temporary, candidate.Path); err != nil {
		return fileResult{kind: fileFailed, reason: fmt.Sprintf("replace the original: %v", err)}
	}
	service.recordSaving(verifyCtx, candidate.Path, candidate.Media.SizeBytes, encoded.SizeBytes)
	if info, err := service.deps.Files.Stat(candidate.Path); err == nil {
		service.cache.store(candidate.Path, info, encoded, nil)
	}
	return fileResult{kind: fileCompressed, bytesAfter: encoded.SizeBytes}
}

func notWorthKeeping(original, encoded Media) (string, bool) {
	switch drift := (encoded.Duration - original.Duration).Abs(); {
	case encoded.Codec != HEVCCodec:
		return fmt.Sprintf("encoded as %s, not %s", encoded.Codec, HEVCCodec), true
	case drift > durationTolerance:
		return fmt.Sprintf("length differs by %s", drift.Round(time.Millisecond)), true
	case encoded.SizeBytes >= original.SizeBytes:
		return fmt.Sprintf("H.265 would be %d bytes, not smaller than %d", encoded.SizeBytes, original.SizeBytes), true
	}
	return "", false
}

func (service *Service) discard(temporary string) {
	if err := service.deps.Files.Remove(temporary); err != nil && !errors.Is(err, fs.ErrNotExist) {
		service.deps.Logger.Warn("leftover temporary encode", "path", temporary, "err", err)
	}
}

func (service *Service) remember(ctx context.Context, skip Skip) {
	if err := service.deps.Skips.Save(ctx, skip); err != nil {
		service.deps.Logger.Warn("will encode this video again next time", "path", skip.Path, "err", err)
	}
}
