package httpapi

import (
	"context"
	"errors"

	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) ListPastedLogins(ctx context.Context, _ gen.ListPastedLoginsRequestObject) (gen.ListPastedLoginsResponseObject, error) {
	if server.deps.Login == nil {
		return gen.ListPastedLogins503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	logins, err := server.deps.Login.PastedLogins(ctx)
	if err != nil {
		return nil, err
	}
	response := make(gen.ListPastedLogins200JSONResponse, len(logins))
	for position, saved := range logins {
		response[position] = toPastedLogin(saved)
	}
	return response, nil
}

func (server apiServer) SavePastedLogin(ctx context.Context, request gen.SavePastedLoginRequestObject) (gen.SavePastedLoginResponseObject, error) {
	if server.deps.Login == nil {
		return gen.SavePastedLogin503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	saved, err := server.deps.Login.SavePastedLogin(ctx, request.Body.Host, request.Body.Token)
	switch {
	case err == nil:
		return gen.SavePastedLogin200JSONResponse(toPastedLogin(saved)), nil
	case login.NeedsUserAction(err):
		return gen.SavePastedLogin400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, login.ErrPortalUnreachable):
		return gen.SavePastedLogin502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	default:
		return nil, err
	}
}

func (server apiServer) ForgetPastedLogin(ctx context.Context, request gen.ForgetPastedLoginRequestObject) (gen.ForgetPastedLoginResponseObject, error) {
	if server.deps.Login == nil {
		return gen.ForgetPastedLogin503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	if err := server.deps.Login.ForgetPastedLogin(ctx, request.Host); err != nil {
		return nil, err
	}
	return gen.ForgetPastedLogin204Response{}, nil
}

func toPastedLogin(saved login.PastedLogin) gen.PastedLogin {
	return gen.PastedLogin{Host: saved.Host, UserName: saved.UserName, SavedAt: saved.SavedAt}
}
