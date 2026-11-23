package httpapi

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vukiman1/udemy-studio/internal/transport/httpapi/gen"
)

func TestCuesFollowTheRequestedLanguageOrder(t *testing.T) {
	app := newTestApp(t)
	cues := decode[[]gen.Cue](t, app.do(t, http.MethodGet, "/api/lectures/"+app.lectureID(t, 1)+"/cues?langs=vi,en", nil))
	want := []gen.Cue{
		{StartMs: 0, EndMs: 1500, Lines: []gen.CueLine{{Language: "vi", Text: "Xin chào"}, {Language: "en", Text: "Hello"}}},
		{StartMs: 1500, EndMs: 3000, Lines: []gen.CueLine{{Language: "vi", Text: "Thế giới"}, {Language: "en", Text: "World"}}},
	}
	if diff := cmp.Diff(want, cues); diff != "" {
		t.Errorf("cues mismatch (-want +got):\n%s", diff)
	}
}

func TestCuesSkipLanguagesTheLectureLacks(t *testing.T) {
	app := newTestApp(t)
	cues := decode[[]gen.Cue](t, app.do(t, http.MethodGet, "/api/lectures/"+app.lectureID(t, 3)+"/cues?langs=vi,en", nil))
	if len(cues) != 2 || len(cues[0].Lines) != 1 || cues[0].Lines[0].Language != "en" {
		t.Errorf("cues = %+v, want english only", cues)
	}
}

func TestCuesOfAnUnknownLectureIsAProblem(t *testing.T) {
	assertProblem(t, newTestApp(t).do(t, http.MethodGet, "/api/lectures/nope/cues", nil), http.StatusNotFound)
}

func TestSubtitlesVTTStacksBothLanguages(t *testing.T) {
	app := newTestApp(t)
	result := app.do(t, http.MethodGet, "/api/lectures/"+app.lectureID(t, 1)+"/subtitles.vtt?langs=vi,en", nil)
	want := "WEBVTT\n\n00:00:00.000 --> 00:00:01.500\nXin chào\nHello\n\n00:00:01.500 --> 00:00:03.000\nThế giới\nWorld\n\n"
	if result.status != http.StatusOK || result.header.Get("Content-Type") != "text/vtt" || string(result.body) != want {
		t.Errorf("status=%d type=%q body=%q", result.status, result.header.Get("Content-Type"), result.body)
	}
}

func TestMalformedQueryParametersAnswerWithAProblem(t *testing.T) {
	app := newTestApp(t)
	result := app.do(t, http.MethodGet, "/api/lectures/"+app.lectureID(t, 1)+"/cues?langs=vi&langs=en", nil)
	assertProblem(t, result, http.StatusBadRequest)
}

func TestOneBrokenSubtitleFileDoesNotHideTheOtherLanguage(t *testing.T) {
	app := newTestAppWith(t, map[string]string{
		"course/01 - Intro/001 Welcome.mp4":    videoBytes,
		"course/01 - Intro/001 Welcome_en.srt": englishSRT,
		"course/01 - Intro/001 Welcome_vi.srt": "not a cue\n00:00:00,000 --> nope\n",
	})
	result := app.do(t, http.MethodGet, "/api/lectures/"+app.lectureID(t, 1)+"/cues?langs=vi,en", nil)
	cues := decode[[]gen.Cue](t, result)
	if result.status != http.StatusOK || len(cues) != 2 || cues[0].Lines[0].Language != "en" {
		t.Errorf("status=%d cues=%+v, want the English track alone", result.status, cues)
	}
}

func TestRepeatedLanguagesAreServedOnce(t *testing.T) {
	app := newTestApp(t)
	cues := decode[[]gen.Cue](t, app.do(t, http.MethodGet, "/api/lectures/"+app.lectureID(t, 1)+"/cues?langs=vi,vi", nil))
	if len(cues) != 2 || len(cues[0].Lines) != 1 {
		t.Errorf("cues = %+v, want one Vietnamese line per cue", cues)
	}
}
