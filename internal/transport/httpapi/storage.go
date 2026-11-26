package httpapi

import (
	"context"
	"errors"

	"github.com/vukiman1/udemy-studio/internal/objectstore"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func (server apiServer) GetStorageSettings(ctx context.Context, _ gen.GetStorageSettingsRequestObject) (gen.GetStorageSettingsResponseObject, error) {
	settings, err := server.storageSettings(ctx)
	if err != nil {
		return nil, err
	}
	return gen.GetStorageSettings200JSONResponse(settings), nil
}

func (server apiServer) SaveStorageSettings(ctx context.Context, request gen.SaveStorageSettingsRequestObject) (gen.SaveStorageSettingsResponseObject, error) {
	err := server.deps.Storage.Save(ctx, storageDraft(*request.Body))
	switch {
	case isBadStorageRequest(err):
		return gen.SaveStorageSettings400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, objectstore.ErrUnreachable):
		return gen.SaveStorageSettings502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	case err != nil:
		return nil, err
	}
	server.followStorage(ctx)
	settings, err := server.storageSettings(ctx)
	if err != nil {
		return nil, err
	}
	return gen.SaveStorageSettings200JSONResponse(settings), nil
}

func (server apiServer) ForgetStorageSettings(ctx context.Context, _ gen.ForgetStorageSettingsRequestObject) (gen.ForgetStorageSettingsResponseObject, error) {
	if err := server.deps.Storage.Forget(ctx); err != nil {
		return nil, err
	}
	server.followStorage(ctx)
	return gen.ForgetStorageSettings204Response{}, nil
}

func (server apiServer) followStorage(ctx context.Context) {
	if server.deps.SyncLibrary == nil {
		return
	}
	if err := server.deps.SyncLibrary(ctx); err != nil {
		server.deps.Logger.Warn("cloud library not refreshed after the storage settings changed", "err", err)
	}
}

func (server apiServer) CheckStorage(ctx context.Context, request gen.CheckStorageRequestObject) (gen.CheckStorageResponseObject, error) {
	check, err := server.deps.Storage.Check(ctx, storageDraft(*request.Body))
	switch {
	case err == nil:
		return gen.CheckStorage200JSONResponse(storageCheck(check)), nil
	case isBadStorageRequest(err):
		return gen.CheckStorage400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err)}, nil
	case errors.Is(err, objectstore.ErrUnreachable):
		return gen.CheckStorage502ApplicationProblemPlusJSONResponse{ProviderUnreachableApplicationProblemPlusJSONResponse: providerUnreachable(err)}, nil
	default:
		return nil, err
	}
}

func (server apiServer) storageSettings(ctx context.Context) (gen.StorageSettings, error) {
	view, err := server.deps.Storage.View(ctx)
	if err != nil || !view.IsConfigured {
		return gen.StorageSettings{}, err
	}
	return gen.StorageSettings{Configured: true, Url: &view.URL, AccessKeyIdHint: &view.AccessKeyIDHint}, nil
}

func storageDraft(body gen.StorageDraft) objectstore.Draft {
	return objectstore.Draft{URL: body.Url, AccessKeyID: valueOf(body.AccessKeyId), SecretAccessKey: valueOf(body.SecretAccessKey)}
}

func storageCheck(check objectstore.Check) gen.StorageCheck {
	return gen.StorageCheck{Access: gen.StorageAccess(check.Access), HasCourse: check.HasIndex}
}

func isBadStorageRequest(err error) bool {
	return errors.Is(err, objectstore.ErrInvalidURL) ||
		errors.Is(err, objectstore.ErrNoCredentials) ||
		errors.Is(err, objectstore.ErrAccessDenied) ||
		errors.Is(err, objectstore.ErrBucketNotFound)
}
