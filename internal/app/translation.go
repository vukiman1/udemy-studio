package app

import (
	"log/slog"
	"os"

	"github.com/vukiman1/udemy-studio/internal/llm"
	"github.com/vukiman1/udemy-studio/internal/translate"
)

func newTranslationQueue(media *os.Root, refresher libraryRefresher, ai *llm.Service, logger *slog.Logger) *translate.Queue {
	return translate.NewQueue(translate.QueueDeps{
		Catalog:      refresher.catalog,
		Media:        media.FS(),
		Files:        media,
		AI:           ai,
		OnTranslated: refresher.afterLibraryChanged,
		Logger:       logger,
	})
}
