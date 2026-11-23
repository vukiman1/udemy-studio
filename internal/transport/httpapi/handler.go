package httpapi

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

type Deps struct {
	Version  string
	Frontend fs.FS
	Catalog  *library.Catalog
	Media    fs.FS
	Progress *progress.Service
	Login    *login.Service
	Jobs     *jobs.Manager
	Logger   *slog.Logger
}

func NewHandler(deps Deps) http.Handler {
	mux := http.NewServeMux()
	api := apiServer{deps: deps}
	gen.HandlerWithOptions(gen.NewStrictHandlerWithOptions(api, nil, strictErrorHandlers(deps.Logger)), gen.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: writeBadRequestProblem,
	})

	files := libraryFiles{catalog: deps.Catalog, media: deps.Media, logger: deps.Logger}
	mux.HandleFunc("GET /media/{lectureId}", files.serveVideo)
	mux.HandleFunc("GET /articles/{lectureId}", files.serveArticle)
	mux.HandleFunc("GET /attachments/{attachmentId}", files.serveAttachment)
	mux.HandleFunc("GET /api/events", jobEventStream{jobs: deps.Jobs}.serve)

	if deps.Frontend != nil {
		mux.Handle("GET /", spaHandler(deps.Frontend))
	}
	return RejectForeignHosts(RejectCrossSiteWrites(mux))
}

type apiServer struct {
	deps Deps
}
