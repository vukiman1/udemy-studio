package httpapi

import (
	"context"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) GetHealth(context.Context, gen.GetHealthRequestObject) (gen.GetHealthResponseObject, error) {
	return gen.GetHealth200JSONResponse{Status: gen.Ok, Version: server.deps.Version}, nil
}
