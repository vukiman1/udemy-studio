package httpapi

import (
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/vukiman1/udemy-studio/internal/jobs"
	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/notes"
	"github.com/vukiman1/udemy-studio/internal/progress"
	"github.com/vukiman1/udemy-studio/internal/search"
	"github.com/vukiman1/udemy-studio/internal/transcode"
	"github.com/vukiman1/udemy-studio/internal/translate"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

type Deps struct {
	Version  string
	Frontend fs.FS
	Catalog  *library.Catalog
	Media    fs.FS
	Progress *progress.Service
	Search   *search.Service
	Notes    *notes.Service
	// Compression is nil when ffmpeg or ffprobe is missing.
	Compression *transcode.Service
	Translation *translate.Queue
	AI          *llm.Service
	Login       *login.Service
	Jobs        *jobs.Manager
	Logger      *slog.Logger
	// PublicHosts are the lowercase hostnames the app is also reached by, such as a tunnel's, besides localhost and IPs.
	PublicHosts []string
}

func NewHandler(deps Deps) http.Handler {
	mux := http.NewServeMux()
	api := apiServer{deps: deps}
	gen.HandlerWithOptions(gen.NewStrictHandlerWithOptions(api, nil, strictErrorHandlers(deps.Logger)), gen.StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: writeBadRequestProblem,
	})

	files := libraryFiles{catalog: deps.Catalog, media: deps.Media, logger: deps.Logger}
	mux.HandleFunc("GET "+mediaPrefix+"{lectureId}", files.serveVideo)
	mux.HandleFunc("GET "+articlesPrefix+"{lectureId}", files.serveArticle)
	mux.HandleFunc("GET "+attachmentsPrefix+"{attachmentId}", files.serveAttachment)
	mux.HandleFunc("GET "+jobEventsPath, jobEventStream{jobs: deps.Jobs}.serve)

	if deps.Frontend != nil {
		mux.Handle("GET /", spaHandler(deps.Frontend))
	}
	return RejectForeignHosts(deps.PublicHosts, RejectCrossSiteWrites(mux))
}

type apiServer struct {
	deps Deps
}
