package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

const downloadsUnavailableDetail = "downloads are not configured: start the server with --python and --python-repo"

func downloadsUnavailable() gen.DownloadsUnavailableApplicationProblemPlusJSONResponse {
	return gen.DownloadsUnavailableApplicationProblemPlusJSONResponse(problem(http.StatusServiceUnavailable, downloadsUnavailableDetail))
}

func (server apiServer) ListChromeProfiles(ctx context.Context, _ gen.ListChromeProfilesRequestObject) (gen.ListChromeProfilesResponseObject, error) {
	if server.deps.Login == nil {
		return gen.ListChromeProfiles503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	profiles, err := server.deps.Login.Profiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("list Chrome profiles: %w", err)
	}
	selected, _, err := server.deps.Login.SelectedProfileDir(ctx)
	if err != nil {
		return nil, err
	}
	response := make(gen.ListChromeProfiles200JSONResponse, 0, len(profiles))
	for _, profile := range profiles {
		response = append(response, toChromeProfile(profile, selected))
	}
	return response, nil
}

func (server apiServer) SelectChromeProfile(ctx context.Context, request gen.SelectChromeProfileRequestObject) (gen.SelectChromeProfileResponseObject, error) {
	if server.deps.Login == nil {
		return gen.SelectChromeProfile503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	err := server.deps.Login.SelectProfile(ctx, request.Body.Dir)
	if errors.Is(err, login.ErrProfileNotFound) {
		return gen.SelectChromeProfile404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound(err.Error())}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.SelectChromeProfile204Response{}, nil
}

// A login problem the learner fixes in Chrome or System Settings is reported in the status, not raised.
func (server apiServer) GetLoginStatus(ctx context.Context, _ gen.GetLoginStatusRequestObject) (gen.GetLoginStatusResponseObject, error) {
	if server.deps.Login == nil {
		return gen.GetLoginStatus503ApplicationProblemPlusJSONResponse{DownloadsUnavailableApplicationProblemPlusJSONResponse: downloadsUnavailable()}, nil
	}
	status, err := server.deps.Login.Status(ctx)
	if err != nil && !login.NeedsUserAction(err) {
		return nil, fmt.Errorf("get login status: %w", err)
	}
	if err != nil {
		server.deps.Logger.Info("login status needs attention", "reason", err)
		problemText := err.Error()
		return gen.GetLoginStatus200JSONResponse{Hosts: []gen.HostLogin{}, Problem: &problemText}, nil
	}
	profile := toChromeProfile(status.Profile, status.Profile.Dir)
	response := gen.GetLoginStatus200JSONResponse{Profile: &profile, Hosts: make([]gen.HostLogin, 0, len(status.Hosts))}
	for _, host := range status.Hosts {
		hostLogin := gen.HostLogin{Host: host.Host, IsExpired: host.IsExpired}
		if !host.ExpiresAt.IsZero() {
			expiresAt := host.ExpiresAt
			hostLogin.ExpiresAt = &expiresAt
		}
		response.Hosts = append(response.Hosts, hostLogin)
	}
	return response, nil
}

func toChromeProfile(profile login.Profile, selectedDir string) gen.ChromeProfile {
	hosts := profile.UdemyHosts
	if hosts == nil {
		hosts = []string{}
	}
	return gen.ChromeProfile{Dir: profile.Dir, Name: profile.Name, UdemyHosts: hosts, Selected: profile.Dir == selectedDir}
}
