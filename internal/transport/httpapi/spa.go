package httpapi

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

const spaIndexFile = "index.html"

func spaHandler(frontend fs.FS) http.Handler {
	files := http.FileServerFS(frontend)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, apiPrefix) {
			http.NotFound(writer, request)
			return
		}
		name := strings.TrimPrefix(path.Clean(request.URL.Path), "/")
		if name == "" {
			files.ServeHTTP(writer, request)
			return
		}
		info, err := fs.Stat(frontend, name)
		if errors.Is(err, fs.ErrNotExist) || (err == nil && info.IsDir()) {
			http.ServeFileFS(writer, request, frontend, spaIndexFile)
			return
		}
		files.ServeHTTP(writer, request)
	})
}
