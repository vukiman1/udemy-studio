package main

import (
	"bufio"
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha1"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

const (
	chromeDataDir           = "Library/Application Support/Google/Chrome"
	localStateFile          = "Local State"
	cookiesFile             = "Cookies"
	keychainService         = "Chrome Safe Storage"
	encryptedValuePrefix    = "v10"
	pbkdf2Salt              = "saltysalt"
	pbkdf2Iterations        = 1003
	aesKeyBytes             = 16
	hostDigestSchemaVersion = 24
	accessTokenCookie       = "access_token"
	udemyHostSuffix         = "udemy.com"
	netscapeFieldCount      = 7
	httpOnlyPrefix          = "#HttpOnly_"
	hashPreviewHexChars     = 12
)

var (
	errProfileNotFound    = errors.New("chrome profile not found")
	errProfileAmbiguous   = errors.New("chrome profile name is ambiguous")
	errUnknownEncryption  = errors.New("unsupported encrypted_value prefix")
	errBadPadding         = errors.New("invalid PKCS7 padding")
	errAccessTokenMissing = errors.New("access_token cookie not found")
)

type localState struct {
	Profile struct {
		InfoCache map[string]struct {
			Name string `json:"name"`
		} `json:"info_cache"`
	} `json:"profile"`
}

type chromeCookie struct {
	host              string
	name              string
	value             []byte
	hostDigestMatched bool
}

func main() {
	profileName := flag.String("profile", "", "Chrome profile display name (ambiguous when names repeat)")
	profileDirFlag := flag.String("profile-dir", "", "Chrome profile directory, e.g. \"Profile 4\"; wins over -profile")
	referenceCookies := flag.String("reference", "", "cookies.txt whose access_token hash is compared")
	flag.Parse()

	if err := run(*profileName, *profileDirFlag, *referenceCookies); err != nil {
		fmt.Fprintln(os.Stderr, "FAIL:", err)
		os.Exit(1)
	}
}

func run(profileName, profileDir, referenceCookies string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}
	dataDir := filepath.Join(home, chromeDataDir)

	if profileDir == "" {
		profileDir, err = findProfileDir(dataDir, profileName)
		if err != nil {
			return err
		}
	}
	fmt.Printf("profile dir %q\n", profileDir)

	snapshotDir, err := os.MkdirTemp("", "chromecookie-*")
	if err != nil {
		return fmt.Errorf("create snapshot dir: %w", err)
	}
	defer os.RemoveAll(snapshotDir)

	snapshotPath, err := snapshotCookiesDB(filepath.Join(dataDir, profileDir), snapshotDir)
	if err != nil {
		return err
	}

	key, err := deriveKey()
	if err != nil {
		return err
	}

	schemaVersion, cookies, err := readUdemyCookies(snapshotPath, key)
	if err != nil {
		return err
	}
	fmt.Printf("meta.version=%d, udemy cookies decrypted=%d\n", schemaVersion, len(cookies))

	digestMatches := 0
	var accessToken *chromeCookie
	for index := range cookies {
		if cookies[index].hostDigestMatched {
			digestMatches++
		}
		if cookies[index].name == accessTokenCookie {
			accessToken = &cookies[index]
		}
	}
	if schemaVersion >= hostDigestSchemaVersion {
		fmt.Printf("host_key SHA-256 prefix verified on %d/%d cookies\n", digestMatches, len(cookies))
	}
	if accessToken == nil {
		return fmt.Errorf("profile %q: %w (log in to Udemy in that profile)", profileDir, errAccessTokenMissing)
	}

	chromeHash := sha256Hex(accessToken.value)
	fmt.Printf("chrome  access_token host=%s len=%d sha256=%s…\n", accessToken.host, len(accessToken.value), chromeHash[:hashPreviewHexChars])

	if referenceCookies == "" {
		return nil
	}
	referenceValue, referenceHost, err := readReferenceAccessToken(referenceCookies)
	if err != nil {
		return err
	}
	referenceHash := sha256Hex(referenceValue)
	fmt.Printf("ref     access_token host=%s len=%d sha256=%s…\n", referenceHost, len(referenceValue), referenceHash[:hashPreviewHexChars])
	if chromeHash != referenceHash {
		return errors.New("access_token hash differs from cookies.txt")
	}
	fmt.Println("PASS: access_token hash matches cookies.txt")
	return nil
}

func findProfileDir(dataDir, profileName string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, localStateFile))
	if err != nil {
		return "", fmt.Errorf("read Local State: %w", err)
	}
	var state localState
	if err := json.Unmarshal(raw, &state); err != nil {
		return "", fmt.Errorf("decode Local State: %w", err)
	}
	var matches []string
	for dir, info := range state.Profile.InfoCache {
		if info.Name == profileName {
			matches = append(matches, dir)
		}
	}
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("%w: %q among %d profiles", errProfileNotFound, profileName, len(state.Profile.InfoCache))
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("%w: %q matches %v, use -profile-dir", errProfileAmbiguous, profileName, matches)
	}
}

func snapshotCookiesDB(profilePath, snapshotDir string) (string, error) {
	for _, suffix := range []string{"", "-wal", "-shm"} {
		source := filepath.Join(profilePath, cookiesFile+suffix)
		if suffix != "" {
			if _, err := os.Stat(source); errors.Is(err, os.ErrNotExist) {
				continue
			}
		}
		if err := copyFile(source, filepath.Join(snapshotDir, cookiesFile+suffix)); err != nil {
			return "", err
		}
	}
	return filepath.Join(snapshotDir, cookiesFile), nil
}

func copyFile(source, destination string) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open %s: %w", source, err)
	}
	defer input.Close()
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create %s: %w", destination, err)
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return fmt.Errorf("copy %s: %w", source, err)
	}
	return output.Close()
}

func deriveKey() ([]byte, error) {
	password, err := exec.Command("security", "find-generic-password", "-w", "-s", keychainService).Output()
	if err != nil {
		return nil, fmt.Errorf("read keychain item %q: %w", keychainService, err)
	}
	return pbkdf2.Key(sha1.New, string(bytes.TrimSpace(password)), []byte(pbkdf2Salt), pbkdf2Iterations, aesKeyBytes)
}

func readUdemyCookies(databasePath string, key []byte) (int, []chromeCookie, error) {
	database, err := sql.Open("sqlite", "file:"+databasePath+"?mode=ro")
	if err != nil {
		return 0, nil, fmt.Errorf("open cookies db: %w", err)
	}
	defer database.Close()

	var versionText string
	if err := database.QueryRow(`SELECT value FROM meta WHERE key = 'version'`).Scan(&versionText); err != nil {
		return 0, nil, fmt.Errorf("read meta.version: %w", err)
	}
	schemaVersion, err := strconv.Atoi(versionText)
	if err != nil {
		return 0, nil, fmt.Errorf("parse meta.version %q: %w", versionText, err)
	}

	rows, err := database.Query(`SELECT host_key, name, value, encrypted_value FROM cookies WHERE host_key LIKE ?`, "%"+udemyHostSuffix)
	if err != nil {
		return 0, nil, fmt.Errorf("query cookies: %w", err)
	}
	defer rows.Close()

	var cookies []chromeCookie
	for rows.Next() {
		var host, name, plainValue string
		var encrypted []byte
		if err := rows.Scan(&host, &name, &plainValue, &encrypted); err != nil {
			return 0, nil, fmt.Errorf("scan cookie row: %w", err)
		}
		if len(encrypted) == 0 {
			cookies = append(cookies, chromeCookie{host: host, name: name, value: []byte(plainValue)})
			continue
		}
		cookie, err := decryptCookie(host, name, encrypted, key, schemaVersion)
		if err != nil {
			return 0, nil, err
		}
		cookies = append(cookies, cookie)
	}
	if err := rows.Err(); err != nil {
		return 0, nil, fmt.Errorf("iterate cookies: %w", err)
	}
	return schemaVersion, cookies, nil
}

func decryptCookie(host, name string, encrypted, key []byte, schemaVersion int) (chromeCookie, error) {
	if !bytes.HasPrefix(encrypted, []byte(encryptedValuePrefix)) {
		return chromeCookie{}, fmt.Errorf("cookie %q on %s: %w %q", name, host, errUnknownEncryption, encrypted[:min(3, len(encrypted))])
	}
	ciphertext := encrypted[len(encryptedValuePrefix):]
	block, err := aes.NewCipher(key)
	if err != nil {
		return chromeCookie{}, fmt.Errorf("create aes cipher: %w", err)
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return chromeCookie{}, fmt.Errorf("cookie %q on %s: ciphertext not block aligned", name, host)
	}
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, bytes.Repeat([]byte{' '}, aes.BlockSize)).CryptBlocks(plaintext, ciphertext)
	plaintext, err = stripPKCS7(plaintext)
	if err != nil {
		return chromeCookie{}, fmt.Errorf("cookie %q on %s: %w", name, host, err)
	}

	cookie := chromeCookie{host: host, name: name, value: plaintext}
	if schemaVersion < hostDigestSchemaVersion {
		return cookie, nil
	}
	hostDigest := sha256.Sum256([]byte(host))
	cookie.hostDigestMatched = bytes.HasPrefix(plaintext, hostDigest[:])
	cookie.value = plaintext[sha256.Size:]
	return cookie, nil
}

func stripPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, errBadPadding
	}
	padding := int(data[len(data)-1])
	if padding == 0 || padding > aes.BlockSize || padding > len(data) {
		return nil, errBadPadding
	}
	paddingStart := len(data) - padding
	if !bytes.Equal(data[paddingStart:], bytes.Repeat([]byte{byte(padding)}, padding)) {
		return nil, errBadPadding
	}
	return data[:paddingStart], nil
}

func readReferenceAccessToken(path string) ([]byte, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, "", fmt.Errorf("open reference cookies: %w", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Split(strings.TrimPrefix(scanner.Text(), httpOnlyPrefix), "\t")
		if len(fields) == netscapeFieldCount && fields[5] == accessTokenCookie {
			return []byte(fields[6]), fields[0], nil
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("read reference cookies: %w", err)
	}
	return nil, "", fmt.Errorf("reference %s: %w", path, errAccessTokenMissing)
}

func sha256Hex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}
