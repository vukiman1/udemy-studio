package app

import (
	"context"
	"log/slog"

	"github.com/vukiman1/udemy-studio/internal/adapter/udemyapi"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

// courseClient talks to the course's portal with the login the app would use: a pasted token, else Chrome's.
func courseClient(ctx context.Context, dataDir, chromeDataDir string, ref udemy.CourseRef, logger *slog.Logger) (*udemyapi.Client, error) {
	database, err := openDatabase(ctx, dataDir)
	if err != nil {
		return nil, err
	}
	defer database.Close()
	box, err := openSettingsBox(dataDir)
	if err != nil {
		return nil, err
	}
	token, err := newLoginService(chromeDataDir, database, box, logger).Token(ctx, ref.Host())
	if err != nil {
		return nil, err
	}
	return visitedClient(ctx, ref, token, logger)
}

func visitedClient(ctx context.Context, ref udemy.CourseRef, token secret.Secret, logger *slog.Logger) (*udemyapi.Client, error) {
	client, err := udemyapi.New(ref.Portal, token, logger)
	if err != nil {
		return nil, err
	}
	if err := client.Visit(ctx); err != nil {
		return nil, err
	}
	return client, nil
}
