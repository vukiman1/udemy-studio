package httpapi

import (
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"mime"
	"net/http"
	"path"

	"github.com/vukiman1/udemy-studio/internal/library"
)

const (
	videoContentType   = "video/mp4"
	articleContentType = "text/html; charset=utf-8"
	binaryContentType  = "application/octet-stream"
	articleSandboxCSP  = "sandbox; default-src 'none'; img-src https: data:; style-src 'unsafe-inline' https:; font-src https: data:"
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
	files.serve(writer, request, lecture.ArticlePath, http.Header{
		"Content-Type":            {articleContentType},
		"Content-Security-Policy": {articleSandboxCSP},
	})
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

func (files libraryFiles) serve(writer http.ResponseWriter, request *http.Request, filePath string, headers http.Header) {
	file, err := files.media.Open(filePath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			files.logger.Warn("library file refused", "path", filePath, "err", err)
		}
		writeProblem(writer, http.StatusNotFound, "file is not available")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		writeProblem(writer, http.StatusNotFound, "file is not available")
		return
	}
	seeker, ok := file.(io.ReadSeeker)
	if !ok {
		files.logger.Error("library file is not seekable", "path", filePath)
		writeProblem(writer, http.StatusInternalServerError, "file cannot be streamed")
		return
	}
	for name, values := range headers {
		writer.Header()[name] = values
	}
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(writer, request, "", info.ModTime(), seeker)
}
