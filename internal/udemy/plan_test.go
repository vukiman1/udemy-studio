package udemy

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

const masterURL = "https://example-portal.udemy.com/assets/77/files/hls/master.m3u8?token=REDACTED"

var errNetworkDown = errors.New("network is down")

type failingPlaylists struct{ err error }

func (playlists failingPlaylists) FetchPlaylist(context.Context, string) (string, error) {
	return "", playlists.err
}

func singleLectureCourse(asset *Asset) []CurriculumItem {
	return []CurriculumItem{
		{Class: "chapter", ID: 1, ObjectIndex: 1, Title: "Intro"},
		{Class: "lecture", ID: 2, ObjectIndex: 2, Title: "Welcome", Asset: asset},
	}
}

func planSingleLecture(t *testing.T, asset *Asset, options PlanOptions, playlists PlaylistFetcher) (LecturePlan, []string) {
	t.Helper()
	plan, err := BuildPlan(t.Context(), Course{Slug: "course"}, singleLectureCourse(asset), options, playlists)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	return plan.Chapters[0].Lectures[0], plan.Warnings
}

func TestLectureWithoutAssetIsPlannedAsUnavailable(t *testing.T) {
	lecture, warnings := planSingleLecture(t, nil, PlanOptions{}, fixturePlaylists{})
	if lecture.Kind != KindUnavailable || len(warnings) != 1 {
		t.Errorf("kind = %q, warnings = %v; want unavailable with one warning", lecture.Kind, warnings)
	}
}

func TestUnavailableHLSMasterFallsBackToMP4WithWarning(t *testing.T) {
	asset := &Asset{AssetType: "Video", StreamURLs: map[string][]VideoSource{videoStreamKey: {
		{Type: "video/mp4", Label: "360", File: "https://mp4-c.udemycdn.com/360.mp4"},
		{Type: hlsMimeType, Label: "Auto", File: masterURL},
	}}}
	lecture, warnings := planSingleLecture(t, asset, PlanOptions{}, failingPlaylists{err: ErrPlaylistUnavailable})
	want := &VideoChoice{Type: SourceMP4, Height: 360, Width: 640, URL: "https://mp4-c.udemycdn.com/360.mp4"}
	if diff := cmp.Diff(want, lecture.Video); diff != "" || len(warnings) != 1 {
		t.Errorf("video (-want +got):\n%s\nwarnings = %v", diff, warnings)
	}
}

func TestAudioOnlySourcesLeaveVideoMissingWithoutWarnings(t *testing.T) {
	asset := &Asset{AssetType: "Video", StreamURLs: map[string][]VideoSource{videoStreamKey: {
		{Type: "audio/mp4", Label: "Audio", File: "https://mp4-c.udemycdn.com/audio.m4a"},
	}}}
	lecture, warnings := planSingleLecture(t, asset, PlanOptions{}, fixturePlaylists{})
	if lecture.Kind != KindVideo || lecture.Video != nil || len(warnings) != 0 {
		t.Errorf("kind = %q, video = %+v, warnings = %v", lecture.Kind, lecture.Video, warnings)
	}
}

func TestHLSFetchFailureOtherThanUnavailableAbortsThePlan(t *testing.T) {
	asset := &Asset{AssetType: "Video", StreamURLs: map[string][]VideoSource{videoStreamKey: {{Type: hlsMimeType, Label: "Auto", File: masterURL}}}}
	_, err := BuildPlan(t.Context(), Course{Slug: "course"}, singleLectureCourse(asset), PlanOptions{}, failingPlaylists{err: errNetworkDown})
	if !errors.Is(err, errNetworkDown) {
		t.Errorf("err = %v, want errNetworkDown", err)
	}
}

func TestCaptionsAreFilteredByLanguage(t *testing.T) {
	asset := &Asset{AssetType: "Video", StreamURLs: map[string][]VideoSource{videoStreamKey: {}}, Captions: []Caption{
		{Class: captionClass, LocaleID: "en_US", URL: "https://vtt-c.udemycdn.com/en.vtt"},
		{Class: captionClass, LocaleID: "vi_VN", URL: "https://vtt-c.udemycdn.com/vi.vtt"},
	}}
	for _, testCase := range []struct {
		language string
		want     []string
	}{
		{"vi", []string{"vi"}},
		{AllCaptionLanguages, []string{"en", "vi"}},
		{"", nil},
	} {
		lecture, _ := planSingleLecture(t, asset, PlanOptions{CaptionLanguage: testCase.language}, fixturePlaylists{})
		var got []string
		for _, caption := range lecture.Captions {
			got = append(got, caption.Language)
		}
		if diff := cmp.Diff(testCase.want, got); diff != "" {
			t.Errorf("language %q (-want +got):\n%s", testCase.language, diff)
		}
	}
}
