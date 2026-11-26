package httpapi

import (
	"context"
	"errors"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

var errNothingToRefresh = errors.New("only the cloud copy reads its library from cloud storage")

func (server apiServer) RefreshLibrary(ctx context.Context, _ gen.RefreshLibraryRequestObject) (gen.RefreshLibraryResponseObject, error) {
	if server.deps.SyncLibrary == nil {
		return gen.RefreshLibrary409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: conflict(errNothingToRefresh)}, nil
	}
	if err := server.deps.SyncLibrary(ctx); err != nil {
		return gen.RefreshLibrary502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	}
	return gen.RefreshLibrary204Response{}, nil
}
