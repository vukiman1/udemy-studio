//go:build !darwin

package app

import (
	"context"
	"log/slog"
)

// keepAwake has nothing to hold off sleep with outside macOS; servers such as a Docker host do not idle-sleep.
func keepAwake(context.Context, *slog.Logger) (stop func()) {
	return func() {}
}
