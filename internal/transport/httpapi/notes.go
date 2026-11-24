package httpapi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"

	"github.com/vukiman1/udemy-studio/internal/notes"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

const notesFileSuffix = " - notes.md"

func (server apiServer) ListLectureNotes(ctx context.Context, request gen.ListLectureNotesRequestObject) (gen.ListLectureNotesResponseObject, error) {
	lecture, ok := server.deps.Catalog.Current().Lecture(request.LectureId)
	if !ok {
		return gen.ListLectureNotes404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound("no lecture " + request.LectureId)}, nil
	}
	lectureNotes, err := server.deps.Notes.ForLecture(ctx, lecture.Key)
	if err != nil {
		return nil, err
	}
	response := make(gen.ListLectureNotes200JSONResponse, len(lectureNotes))
	for position, note := range lectureNotes {
		response[position] = toNote(note)
	}
	return response, nil
}

func (server apiServer) AddLectureNote(ctx context.Context, request gen.AddLectureNoteRequestObject) (gen.AddLectureNoteResponseObject, error) {
	lecture, ok := server.deps.Catalog.Current().Lecture(request.LectureId)
	if !ok {
		return gen.AddLectureNote404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound("no lecture " + request.LectureId)}, nil
	}
	note, err := server.deps.Notes.Add(ctx, lecture.Key, request.Body.AtSeconds, request.Body.Body)
	if errors.Is(err, notes.ErrInvalidNote) {
		return gen.AddLectureNote400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.AddLectureNote201JSONResponse(toNote(note)), nil
}

func (server apiServer) EditNote(ctx context.Context, request gen.EditNoteRequestObject) (gen.EditNoteResponseObject, error) {
	note, err := server.deps.Notes.Edit(ctx, request.NoteId, request.Body.Body)
	switch {
	case errors.Is(err, notes.ErrInvalidNote):
		return gen.EditNote400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, notes.ErrNotFound):
		return gen.EditNote404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: noSuchNote(request.NoteId)}, nil
	case err != nil:
		return nil, err
	}
	return gen.EditNote200JSONResponse(toNote(note)), nil
}

func (server apiServer) DeleteNote(ctx context.Context, request gen.DeleteNoteRequestObject) (gen.DeleteNoteResponseObject, error) {
	err := server.deps.Notes.Delete(ctx, request.NoteId)
	if errors.Is(err, notes.ErrNotFound) {
		return gen.DeleteNote404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: noSuchNote(request.NoteId)}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.DeleteNote204Response{}, nil
}

func (server apiServer) ExportCourseNotes(ctx context.Context, request gen.ExportCourseNotesRequestObject) (gen.ExportCourseNotesResponseObject, error) {
	course, ok := server.deps.Catalog.Current().Course(request.CourseId)
	if !ok {
		return gen.ExportCourseNotes404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound("no course " + request.CourseId)}, nil
	}
	byLecture, err := server.deps.Notes.ByLecture(ctx)
	if err != nil {
		return nil, err
	}
	var document bytes.Buffer
	if err := notes.WriteMarkdown(&document, course, byLecture); err != nil {
		return nil, err
	}
	return gen.ExportCourseNotes200TextmarkdownCharsetUtf8Response{
		Body:          &document,
		ContentLength: int64(document.Len()),
		Headers: gen.ExportCourseNotes200ResponseHeaders{
			ContentDisposition: mime.FormatMediaType("attachment", map[string]string{"filename": course.Title + notesFileSuffix}),
		},
	}, nil
}

func toNote(note notes.Note) gen.Note {
	return gen.Note{
		Id:        note.ID,
		AtSeconds: note.AtSeconds,
		Body:      note.Body,
		CreatedAt: note.CreatedAt,
		UpdatedAt: note.UpdatedAt,
	}
}

func noSuchNote(id int64) gen.NotFoundApplicationProblemPlusJSONResponse {
	return notFound(fmt.Sprintf("no note %d", id))
}
