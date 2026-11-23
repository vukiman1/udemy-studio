package httpapi

import (
	"bytes"
	"context"
	"fmt"
	"slices"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/subtitle"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) GetLectureCues(ctx context.Context, request gen.GetLectureCuesRequestObject) (gen.GetLectureCuesResponseObject, error) {
	lecture, ok := server.deps.Library.Lecture(request.LectureId)
	if !ok {
		return gen.GetLectureCues404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound("no lecture " + request.LectureId)}, nil
	}
	merged, err := server.mergedCues(ctx, lecture, request.Params.Langs)
	if err != nil {
		return nil, err
	}
	cues := make(gen.GetLectureCues200JSONResponse, 0, len(merged))
	for _, cue := range merged {
		lines := make([]gen.CueLine, 0, len(cue.Lines))
		for _, line := range cue.Lines {
			lines = append(lines, gen.CueLine{Language: line.Language, Text: line.Text})
		}
		cues = append(cues, gen.Cue{StartMs: int(cue.Start.Milliseconds()), EndMs: int(cue.End.Milliseconds()), Lines: lines})
	}
	return cues, nil
}

func (server apiServer) GetLectureSubtitlesVtt(ctx context.Context, request gen.GetLectureSubtitlesVttRequestObject) (gen.GetLectureSubtitlesVttResponseObject, error) {
	lecture, ok := server.deps.Library.Lecture(request.LectureId)
	if !ok {
		return gen.GetLectureSubtitlesVtt404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound("no lecture " + request.LectureId)}, nil
	}
	merged, err := server.mergedCues(ctx, lecture, request.Params.Langs)
	if err != nil {
		return nil, err
	}
	var document bytes.Buffer
	if err := subtitle.WriteVTT(&document, merged); err != nil {
		return nil, fmt.Errorf("render vtt for lecture %s: %w", lecture.ID, err)
	}
	return gen.GetLectureSubtitlesVtt200TextvttResponse{Body: &document, ContentLength: int64(document.Len())}, nil
}

func (server apiServer) mergedCues(ctx context.Context, lecture library.Lecture, requested *gen.Languages) ([]subtitle.MergedCue, error) {
	files := selectSubtitles(lecture.Subtitles, requested)
	tracks := make([]subtitle.Track, 0, len(files))
	for _, file := range files {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		cues, err := server.parseSubtitle(file.Path)
		if err != nil {
			server.deps.Logger.Warn("skipping unreadable subtitle", "lecture", lecture.ID, "language", file.Language, "path", file.Path, "err", err)
			continue
		}
		tracks = append(tracks, subtitle.Track{Language: file.Language, Cues: cues})
	}
	return subtitle.Merge(tracks), nil
}

func (server apiServer) parseSubtitle(path string) ([]subtitle.Cue, error) {
	file, err := server.deps.Media.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer file.Close()
	return subtitle.ParseSRT(file)
}

func selectSubtitles(available []library.SubtitleFile, requested *gen.Languages) []library.SubtitleFile {
	if requested == nil {
		return available
	}
	var selected []library.SubtitleFile
	for _, language := range *requested {
		if slices.ContainsFunc(selected, func(file library.SubtitleFile) bool { return file.Language == language }) {
			continue
		}
		for _, file := range available {
			if file.Language == language {
				selected = append(selected, file)
			}
		}
	}
	return selected
}
