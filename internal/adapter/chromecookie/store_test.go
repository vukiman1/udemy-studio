package chromecookie

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/vukiman1/udemy-studio/internal/login"
)

const (
	portalHost   = "acme.udemy.com"
	schemaV24    = 24
	localStateV1 = `{"profile":{"info_cache":{"Profile 4":{"name":"work"},"Profile 13":{"name":"work"},"Default":{"name":"home"}}}}`
)

var tokenExpiry = time.Date(2027, 1, 2, 3, 4, 5, 0, time.UTC)

type fixedPassword struct{ calls int }

func (source *fixedPassword) SafeStoragePassword(context.Context) ([]byte, error) {
	source.calls++
	return []byte(testPassword), nil
}

type cookieRow struct {
	hostKey string
	name    string
	value   string
}

func writeCookiesDB(t *testing.T, profilePath string, rows ...cookieRow) {
	t.Helper()
	if err := os.MkdirAll(profilePath, 0o750); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	database, err := sql.Open(sqliteDriver, filepath.Join(profilePath, cookiesFile))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer database.Close()
	statements := []string{
		`CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT)`,
		`INSERT INTO meta VALUES ('version', '24')`,
		`CREATE TABLE cookies (host_key TEXT, name TEXT, value TEXT, encrypted_value BLOB, expires_utc INTEGER)`,
	}
	for _, statement := range statements {
		if _, err := database.ExecContext(t.Context(), statement); err != nil {
			t.Fatalf("exec %q: %v", statement, err)
		}
	}
	expiresMicros := tokenExpiry.UnixMicro() + chromeToUnixEpochMicros
	for _, row := range rows {
		encrypted := encryptForTest(t, row.value, row.hostKey, schemaV24)
		if _, err := database.ExecContext(t.Context(), `INSERT INTO cookies VALUES (?, ?, '', ?, ?)`, row.hostKey, row.name, encrypted, expiresMicros); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}
}

func newChromeDir(t *testing.T) string {
	t.Helper()
	dataDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dataDir, localStateFile), []byte(localStateV1), 0o600); err != nil {
		t.Fatalf("write Local State: %v", err)
	}
	writeCookiesDB(t, filepath.Join(dataDir, "Profile 4"),
		cookieRow{hostKey: portalHost, name: accessTokenCookie, value: "work-token"},
		cookieRow{hostKey: ".udemy.com", name: "ud_country_code", value: "VN"},
	)
	writeCookiesDB(t, filepath.Join(dataDir, "Profile 13"), cookieRow{hostKey: ".google.com", name: "SID", value: "x"})
	return dataDir
}

func TestProfilesListEveryProfileWithItsUdemyLogins(t *testing.T) {
	profiles, err := NewStore(newChromeDir(t), &fixedPassword{}).Profiles(t.Context())
	if err != nil {
		t.Fatalf("Profiles: %v", err)
	}
	want := []login.Profile{
		{Dir: "Default", Name: "home"},
		{Dir: "Profile 13", Name: "work"},
		{Dir: "Profile 4", Name: "work", UdemyHosts: []string{portalHost}},
	}
	if diff := cmp.Diff(want, profiles); diff != "" {
		t.Errorf("profiles mismatch (-want +got):\n%s", diff)
	}
}

func TestAccessTokenDecryptsTheCookieOfThatPortal(t *testing.T) {
	passwords := &fixedPassword{}
	store := NewStore(newChromeDir(t), passwords)
	for range 2 {
		token, err := store.AccessToken(t.Context(), "Profile 4", portalHost)
		if err != nil {
			t.Fatalf("AccessToken: %v", err)
		}
		if token.Value.Reveal() != "work-token" || !token.ExpiresAt.Equal(tokenExpiry) {
			t.Errorf("token = %q expiring %v", token.Value.Reveal(), token.ExpiresAt)
		}
	}
	if passwords.calls != 1 {
		t.Errorf("keychain asked %d times, want once", passwords.calls)
	}
}

func TestAccessTokenFailuresAreNamed(t *testing.T) {
	dataDir := newChromeDir(t)
	store := NewStore(dataDir, &fixedPassword{})
	cases := []struct {
		name       string
		profileDir string
		host       string
		want       error
	}{
		{name: "profile without a Udemy login", profileDir: "Profile 13", host: portalHost, want: login.ErrNotLoggedIn},
		{name: "another portal", profileDir: "Profile 4", host: "www.udemy.com", want: login.ErrNotLoggedIn},
		{name: "path traversal in the profile", profileDir: "../Profile 4", host: portalHost, want: login.ErrProfileNotFound},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if _, err := store.AccessToken(t.Context(), testCase.profileDir, testCase.host); !errors.Is(err, testCase.want) {
				t.Errorf("error = %v, want %v", err, testCase.want)
			}
		})
	}
}

func TestUnreadableChromeDataAsksForDiskAccess(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores file permissions")
	}
	dataDir := newChromeDir(t)
	cookies := filepath.Join(dataDir, "Profile 4", cookiesFile)
	if err := os.Chmod(cookies, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(cookies, 0o600) })
	if _, err := NewStore(dataDir, &fixedPassword{}).AccessToken(t.Context(), "Profile 4", portalHost); !errors.Is(err, login.ErrNoDiskAccess) {
		t.Errorf("error = %v, want ErrNoDiskAccess", err)
	}
}

func TestAccessTokenFallsBackToTheParentDomainCookie(t *testing.T) {
	dataDir := newChromeDir(t)
	writeCookiesDB(t, filepath.Join(dataDir, "Default"), cookieRow{hostKey: ".udemy.com", name: accessTokenCookie, value: "parent-token"})
	writeCookiesDB(t, filepath.Join(dataDir, "Profile 99"),
		cookieRow{hostKey: ".udemy.com", name: accessTokenCookie, value: "parent-token"},
		cookieRow{hostKey: "www.udemy.com", name: accessTokenCookie, value: "exact-token"},
	)
	store := NewStore(dataDir, &fixedPassword{})
	cases := map[string]string{"Default": "parent-token", "Profile 99": "exact-token"}
	for profileDir, want := range cases {
		token, err := store.AccessToken(t.Context(), profileDir, "www.udemy.com")
		if err != nil || token.Value.Reveal() != want {
			t.Errorf("%s: token = %q, %v; want %q", profileDir, token.Value.Reveal(), err, want)
		}
	}
}

func TestProfileReadsOnlyTheRequestedProfile(t *testing.T) {
	dataDir := newChromeDir(t)
	if err := os.Chmod(filepath.Join(dataDir, "Profile 13", cookiesFile), 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(dataDir, "Profile 13", cookiesFile), 0o600) })
	profile, err := NewStore(dataDir, &fixedPassword{}).Profile(t.Context(), "Profile 4")
	if err != nil {
		t.Fatalf("Profile: %v (it must not touch other profiles)", err)
	}
	if profile.Name != "work" || len(profile.UdemyHosts) != 1 {
		t.Errorf("profile = %+v", profile)
	}
	if _, err := NewStore(dataDir, &fixedPassword{}).Profile(t.Context(), "Profile 77"); !errors.Is(err, login.ErrProfileNotFound) {
		t.Errorf("unknown profile error = %v", err)
	}
}

func TestAMachineWithoutChromeHasNoProfiles(t *testing.T) {
	profiles, err := NewStore(filepath.Join(t.TempDir(), "no-chrome-here"), nil).Profiles(t.Context())
	if err != nil || len(profiles) != 0 {
		t.Errorf("Profiles = %+v, %v; want none and no error", profiles, err)
	}
}
