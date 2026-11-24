package sqlite

import (
	"log/slog"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vukiman1/udemy-studio/internal/search"
)

func openTestSearch(t *testing.T) *search.Service {
	t.Helper()
	database, err := Open(t.Context(), filepath.Join(t.TempDir(), "app.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	index := NewCueIndex(database)
	err = index.Replace(t.Context(), []search.Entry{
		search.NewEntry("udemy-lecture:1", "vi", 61500*time.Millisecond, "Đây là đường truyền kết nối TCP"),
		search.NewEntry("udemy-lecture:2", "en", 2*time.Second, "WebSockets keep the connection open"),
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}
	return search.NewService(index, slog.New(slog.DiscardHandler))
}

func TestSubtitleSearchIgnoresDiacriticsIncludingDAndMatchesTheLastWordAsAPrefix(t *testing.T) {
	for _, query := range []string{"duong truyen ket no", "đường truyền kết nối", "DAY la"} {
		hits, err := openTestSearch(t).Search(t.Context(), query)
		if err != nil || len(hits) != 1 {
			t.Fatalf("Search(%q) = %+v, %v; want the Vietnamese cue", query, hits, err)
		}
		if hit := hits[0]; hit.LectureKey != "udemy-lecture:1" || hit.Language != "vi" || hit.Start != 61500*time.Millisecond {
			t.Errorf("Search(%q) hit = %+v", query, hit)
		}
	}
}

func TestSubtitleSearchHighlightsTheOriginalText(t *testing.T) {
	hits, err := openTestSearch(t).Search(t.Context(), "duong")
	if err != nil || len(hits) != 1 {
		t.Fatalf("Search = %+v, %v", hits, err)
	}
	want := "Đây là " + search.HighlightStart + "đường" + search.HighlightEnd + " truyền kết nối TCP"
	if hits[0].Snippet != want {
		t.Errorf("snippet = %q, want %q", hits[0].Snippet, want)
	}
}

func TestSubtitleSearchTreatsTypedSyntaxAsText(t *testing.T) {
	service := openTestSearch(t)
	for _, query := range []string{`NEAR(`, `"`, "-connection", "AND", "web* OR"} {
		if _, err := service.Search(t.Context(), query); err != nil {
			t.Errorf("Search(%q): %v", query, err)
		}
	}
	if hits, _ := service.Search(t.Context(), strings.Repeat(" ", 3)); hits != nil {
		t.Errorf("blank query hits = %+v", hits)
	}
}
