package udemy

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseMasterPlaylistMatchesPythonM3U8(t *testing.T) {
	for _, goldenPath := range []string{"udemy/plan/hls_golden.json", "udemy/plan/edge_hls_golden.json"} {
		for masterURL, want := range readJSON[map[string][]HLSVariant](t, goldenPath) {
			playlist, err := fixturePlaylists{}.FetchPlaylist(t.Context(), masterURL)
			if err != nil {
				t.Fatalf("fixture for %s: %v", masterURL, err)
			}
			if diff := cmp.Diff(want, ParseMasterPlaylist(playlist)); diff != "" {
				t.Errorf("variants of %s differ from python (-python +go):\n%s", masterURL, diff)
			}
		}
	}
}

func TestHLSVariantURIIsResolvedAgainstTheMaster(t *testing.T) {
	got := resolveReference("https://example-portal.udemy.com/assets/1/files/hls/master.m3u8?token=REDACTED", "720/index.m3u8")
	if want := "https://example-portal.udemy.com/assets/1/files/hls/720/index.m3u8"; got != want {
		t.Errorf("resolveReference = %q, want %q", got, want)
	}
}
