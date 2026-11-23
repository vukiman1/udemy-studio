package chromecookie

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/vukiman1/udemy-studio/internal/login"
	"github.com/vukiman1/udemy-studio/internal/platform/secret"
)

const (
	accessTokenCookie = "access_token"
	udemyHostPattern  = "%udemy.com"
	sqliteDriver      = "sqlite"
)

// Chrome stores expiry as microseconds since 1601-01-01 UTC, beyond what time.Duration can hold.
const chromeToUnixEpochMicros = 11_644_473_600_000_000

type PasswordSource interface {
	SafeStoragePassword(ctx context.Context) ([]byte, error)
}

type Store struct {
	dataDir   string
	passwords PasswordSource
	keyMutex  sync.Mutex
	key       []byte
}

func NewStore(dataDir string, passwords PasswordSource) *Store {
	return &Store{dataDir: dataDir, passwords: passwords}
}

func (store *Store) Profiles(ctx context.Context) ([]login.Profile, error) {
	entries, err := readProfiles(store.dataDir)
	if err != nil {
		return nil, err
	}
	profiles := make([]login.Profile, 0, len(entries))
	for _, entry := range entries {
		hosts, err := store.loggedInHosts(ctx, entry.dir)
		if err != nil {
			return nil, fmt.Errorf("read Udemy logins of profile %q: %w", entry.dir, err)
		}
		profiles = append(profiles, login.Profile{Dir: entry.dir, Name: entry.name, UdemyHosts: hosts})
	}
	return profiles, nil
}

func (store *Store) Profile(ctx context.Context, profileDir string) (login.Profile, error) {
	entries, err := readProfiles(store.dataDir)
	if err != nil {
		return login.Profile{}, err
	}
	position := slices.IndexFunc(entries, func(entry profileEntry) bool { return entry.dir == profileDir })
	if position < 0 {
		return login.Profile{}, fmt.Errorf("%w: %q", login.ErrProfileNotFound, profileDir)
	}
	hosts, err := store.loggedInHosts(ctx, profileDir)
	if err != nil {
		return login.Profile{}, fmt.Errorf("read Udemy logins of profile %q: %w", profileDir, err)
	}
	return login.Profile{Dir: profileDir, Name: entries[position].name, UdemyHosts: hosts}, nil
}

func (store *Store) AccessToken(ctx context.Context, profileDir, host string) (login.Token, error) {
	profilePath, err := store.profilePath(profileDir)
	if err != nil {
		return login.Token{}, err
	}
	database, cleanup, err := openSnapshot(ctx, profilePath)
	if err != nil {
		return login.Token{}, err
	}
	defer cleanup()

	schemaVersion, err := readSchemaVersion(ctx, database)
	if err != nil {
		return login.Token{}, err
	}
	var hostKey string
	var encrypted []byte
	var expiresMicros int64
	domains := cookieDomainsFor(host)
	arguments := []any{accessTokenCookie}
	for _, domain := range domains {
		arguments = append(arguments, domain)
	}
	query := `SELECT host_key, encrypted_value, expires_utc FROM cookies
		WHERE name = ? AND host_key IN (?` + strings.Repeat(", ?", len(domains)-1) + `)
		ORDER BY length(host_key) DESC, expires_utc DESC LIMIT 1`
	err = database.QueryRowContext(ctx, query, arguments...).Scan(&hostKey, &encrypted, &expiresMicros)
	if errors.Is(err, sql.ErrNoRows) {
		return login.Token{}, login.ErrNotLoggedIn
	}
	if err != nil {
		return login.Token{}, fmt.Errorf("query access_token for %s: %w", host, err)
	}
	key, err := store.cookieKey(ctx)
	if err != nil {
		return login.Token{}, err
	}
	value, err := decryptValue(encrypted, key, hostKey, schemaVersion)
	if err != nil {
		return login.Token{}, fmt.Errorf("decrypt access_token for %s: %w", host, err)
	}
	return login.Token{Value: secret.New(string(value)), ExpiresAt: chromeTime(expiresMicros)}, nil
}

func (store *Store) loggedInHosts(ctx context.Context, profileDir string) ([]string, error) {
	profilePath, err := store.profilePath(profileDir)
	if err != nil {
		return nil, err
	}
	database, cleanup, err := openSnapshot(ctx, profilePath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer cleanup()
	rows, err := database.QueryContext(ctx,
		`SELECT DISTINCT ltrim(host_key, '.') FROM cookies WHERE name = ? AND host_key LIKE ? ORDER BY 1`,
		accessTokenCookie, udemyHostPattern)
	if err != nil {
		return nil, fmt.Errorf("query Udemy logins: %w", err)
	}
	defer rows.Close()
	var hosts []string
	for rows.Next() {
		var host string
		if err := rows.Scan(&host); err != nil {
			return nil, fmt.Errorf("scan Udemy login host: %w", err)
		}
		hosts = append(hosts, host)
	}
	return hosts, rows.Err()
}

// A cookie set for a parent domain (.udemy.com) is sent to www.udemy.com too; the most specific one wins.
func cookieDomainsFor(host string) []string {
	domains := []string{host, "." + host}
	labels := strings.Split(host, ".")
	for index := 1; index < len(labels)-1; index++ {
		domains = append(domains, "."+strings.Join(labels[index:], "."))
	}
	return domains
}

func (store *Store) profilePath(profileDir string) (string, error) {
	if profileDir == "" || filepath.Base(profileDir) != profileDir {
		return "", fmt.Errorf("%w: %q", login.ErrProfileNotFound, profileDir)
	}
	return filepath.Join(store.dataDir, profileDir), nil
}

func (store *Store) cookieKey(ctx context.Context) ([]byte, error) {
	store.keyMutex.Lock()
	defer store.keyMutex.Unlock()
	if store.key != nil {
		return store.key, nil
	}
	password, err := store.passwords.SafeStoragePassword(ctx)
	if err != nil {
		return nil, err
	}
	key, err := deriveKey(password)
	if err != nil {
		return nil, err
	}
	store.key = key
	return key, nil
}

func readSchemaVersion(ctx context.Context, database *sql.DB) (int, error) {
	var versionText string
	if err := database.QueryRowContext(ctx, `SELECT value FROM meta WHERE key = 'version'`).Scan(&versionText); err != nil {
		return 0, fmt.Errorf("read cookies schema version: %w", err)
	}
	version, err := strconv.Atoi(versionText)
	if err != nil {
		return 0, fmt.Errorf("parse cookies schema version %q: %w", versionText, err)
	}
	return version, nil
}

func chromeTime(microseconds int64) time.Time {
	if microseconds == 0 {
		return time.Time{}
	}
	return time.UnixMicro(microseconds - chromeToUnixEpochMicros).UTC()
}
