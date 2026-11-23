package app

import (
	"context"
	"log/slog"

	"github.com/vukiman1/udemy-studio/internal/adapter/udemyapi"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

// courseClient talks to the course's portal with the login of the Chrome profile selected in the app.
func courseClient(ctx context.Context, dataDir, chromeDataDir string, ref udemy.CourseRef, logger *slog.Logger) (*udemyapi.Client, error) {
	database, err := openDatabase(ctx, dataDir)
	if err != nil {
		return nil, err
	}
	defer database.Close()
	token, err := newLoginService(chromeDataDir, database).Token(ctx, ref.Host())
	if err != nil {
		return nil, err
	}
	client, err := udemyapi.New(ref.Portal, token, logger)
	if err != nil {
		return nil, err
	}
	if err := client.Visit(ctx); err != nil {
		return nil, err
	}
	return client, nil
}
