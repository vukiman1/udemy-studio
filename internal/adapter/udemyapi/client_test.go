package udemyapi

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestServiceUnavailableIsRetried(t *testing.T) {
	fake := &fakeTransport{}
	target := testPortalURL + "/api-2.0/x"
	fake.reply(target, http.StatusServiceUnavailable, "busy")
	fake.reply(target, http.StatusBadGateway, "busy")
	fake.reply(target, http.StatusOK, `{"value": 1}`)
	var got struct{ Value int }
	if err := newTestClient(t, fake).getJSON(t.Context(), target, &got); err != nil || got.Value != 1 {
		t.Fatalf("getJSON = %+v, %v", got, err)
	}
	if attempts := fake.requestedMatching("/api-2.0/x"); attempts != 3 {
		t.Errorf("attempts = %d, want 3", attempts)
	}
}

func TestCloudflareChallengeStopsImmediately(t *testing.T) {
	fake := &fakeTransport{}
	target := testPortalURL + "/api-2.0/x"
	fake.reply(target, http.StatusForbidden, "<html><title>Just a moment...</title></html>")
	if err := newTestClient(t, fake).getJSON(t.Context(), target, &struct{}{}); !errors.Is(err, ErrChallenged) {
		t.Errorf("err = %v, want ErrChallenged", err)
	}
}

func TestErrorsNameTheStatusButNotTheQuery(t *testing.T) {
	fake := &fakeTransport{}
	target := testPortalURL + "/assets/1/master.m3u8?token=secret-token"
	fake.reply(target, http.StatusForbidden, `{"detail": "You do not have permission."}`)
	_, err := newTestClient(t, fake).FetchPlaylist(t.Context(), target)
	if err == nil || strings.Contains(err.Error(), "secret-token") || !strings.Contains(err.Error(), "You do not have permission") {
		t.Errorf("err = %v", err)
	}
}

func TestCurriculumFollowsEveryPage(t *testing.T) {
	fake := &fakeTransport{}
	first := testPortalURL + "/api-2.0/courses/7/subscriber-curriculum-items/?" + curriculumParams.Encode()
	second := testPortalURL + "/api-2.0/courses/7/subscriber-curriculum-items/?page=2"
	fake.reply(first, http.StatusOK, `{"next": "`+second+`", "results": [{"_class": "chapter", "id": 1, "title": "Intro"}]}`)
	fake.reply(second, http.StatusOK, `{"next": null, "results": [{"_class": "lecture", "id": 2, "title": "Welcome"}]}`)
	items, err := newTestClient(t, fake).Curriculum(t.Context(), 7)
	if err != nil || len(items) != 2 || items[1].Title != "Welcome" {
		t.Errorf("Curriculum = %+v, %v", items, err)
	}
}

func TestChallengeMarkerInASuccessfulPageIsNotAChallenge(t *testing.T) {
	fake := &fakeTransport{}
	target := testPortalURL + "/course/backend-basics/learn/"
	page := `<html><script src="/cdn-cgi/challenge-platform/scripts/jsd/main.js"></script></html>`
	fake.reply(target, http.StatusOK, page)
	if got, err := newTestClient(t, fake).getText(t.Context(), target); err != nil || got != page {
		t.Errorf("getText = %q, %v", got, err)
	}
}
