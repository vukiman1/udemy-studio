package httpapi

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"
	"time"

	"github.com/vukiman1/udemy-studio/internal/library"
)

const (
	videoContentType   = "video/mp4"
	articleContentType = "text/html; charset=utf-8"
	binaryContentType  = "application/octet-stream"
	coverContentType   = "image/jpeg"
	articleSandboxCSP  = "sandbox; default-src 'none'; img-src https: data:; style-src 'unsafe-inline' https:; font-src https: data:"
	// A quiz page runs its own inline scripts, still in a sandbox with an opaque origin and nothing it can fetch.
	quizSandboxCSP  = "sandbox allow-scripts; default-src 'none'; script-src 'unsafe-inline'; img-src https: data:; style-src 'unsafe-inline' https:; font-src https: data:"
	maxArticleBytes = 32 << 20
)

type libraryFiles struct {
	catalog *library.Catalog
	media   fs.FS
	logger  *slog.Logger
}

func (files libraryFiles) serveVideo(writer http.ResponseWriter, request *http.Request) {
	lecture, ok := files.catalog.Current().Lecture(request.PathValue("lectureId"))
	if !ok || lecture.VideoPath == "" {
		writeProblem(writer, http.StatusNotFound, "no video for lecture "+request.PathValue("lectureId"))
		return
	}
	files.serve(writer, request, lecture.VideoPath, http.Header{"Content-Type": {videoContentType}})
}

func (files libraryFiles) serveArticle(writer http.ResponseWriter, request *http.Request) {
	lecture, ok := files.catalog.Current().Lecture(request.PathValue("lectureId"))
	if !ok || lecture.ArticlePath == "" {
		writeProblem(writer, http.StatusNotFound, "no article for lecture "+request.PathValue("lectureId"))
		return
	}
	policy, themeCSS := articleSandboxCSP, articleThemeCSS
	if lecture.Kind == library.LectureKindQuiz {
		policy, themeCSS = quizSandboxCSP, quizThemeCSS
	}
	file, _, ok := files.open(writer, lecture.ArticlePath)
	if !ok {
		return
	}
	defer file.Close()
	page, err := io.ReadAll(io.LimitReader(file, maxArticleBytes+1))
	if err != nil || len(page) > maxArticleBytes {
		files.logger.Error("article page unreadable", "path", lecture.ArticlePath, "bytes", len(page), "err", err)
		writeProblem(writer, http.StatusInternalServerError, "article cannot be read")
		return
	}
	setFileHeaders(writer, http.Header{
		"Content-Type":            {articleContentType},
		"Content-Security-Policy": {policy},
	})
	http.ServeContent(writer, request, "", time.Time{}, bytes.NewReader(withPageTheme(page, themeCSS)))
}

func (files libraryFiles) serveAttachment(writer http.ResponseWriter, request *http.Request) {
	attachment, ok := files.catalog.Current().Attachment(request.PathValue("attachmentId"))
	if !ok {
		writeProblem(writer, http.StatusNotFound, "no attachment "+request.PathValue("attachmentId"))
		return
	}
	contentType := mime.TypeByExtension(path.Ext(attachment.Name))
	if contentType == "" {
		contentType = binaryContentType
	}
	files.serve(writer, request, attachment.Path, http.Header{
		"Content-Type":        {contentType},
		"Content-Disposition": {mime.FormatMediaType("attachment", map[string]string{"filename": attachment.Name})},
	})
}

func (files libraryFiles) serveCover(writer http.ResponseWriter, request *http.Request) {
	course, ok := files.catalog.Current().Course(request.PathValue("courseId"))
	if !ok || course.CoverPath == "" {
		writeProblem(writer, http.StatusNotFound, "no cover for course "+request.PathValue("courseId"))
		return
	}
	files.serve(writer, request, course.CoverPath, http.Header{"Content-Type": {coverContentType}})
}

func (files libraryFiles) serve(writer http.ResponseWriter, request *http.Request, filePath string, headers http.Header) {
	file, info, ok := files.open(writer, filePath)
	if !ok {
		return
	}
	defer file.Close()
	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		files.logger.Error("library file is not seekable", "path", filePath)
		writeProblem(writer, http.StatusInternalServerError, "file cannot be streamed")
		return
	}
	setFileHeaders(writer, headers)
	http.ServeContent(writer, request, "", info.ModTime(), seeker)
}

func (files libraryFiles) open(writer http.ResponseWriter, filePath string) (fs.File, fs.FileInfo, bool) {
	file, err := files.media.Open(filePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			files.logger.Warn("library file refused", "path", filePath, "err", err)
		}
		writeProblem(writer, http.StatusNotFound, "file is not available")
		return nil, nil, false
	}
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		_ = file.Close()
		writeProblem(writer, http.StatusNotFound, "file is not available")
		return nil, nil, false
	}
	return file, info, true
}

func setFileHeaders(writer http.ResponseWriter, headers http.Header) {
	for name, values := range headers {
		writer.Header()[name] = values
	}
	writer.Header().Set("X-Content-Type-Options", "nosniff")
}
