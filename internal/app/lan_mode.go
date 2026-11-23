package app

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/vukiman1/udemy-studio/internal/platform/terminalqr"
	"github.com/vukiman1/udemy-studio/internal/transport/httpapi"
)

const (
	lanAccessCodeFileName = "lan-access-code"
	accessCodeBytes       = 8
	accessCodeFileMode    = 0o600
	allInterfaces         = "0.0.0.0"
)

var accessCodeEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

func enableLANMode(address string, handler http.Handler, dataDir string) (string, http.Handler, string, error) {
	lanAddress, err := lanListenAddress(address)
	if err != nil {
		return "", nil, "", err
	}
	accessCode, err := loadOrCreateAccessCode(filepath.Join(dataDir, lanAccessCodeFileName))
	if err != nil {
		return "", nil, "", err
	}
	return lanAddress, httpapi.RequireLANAccess(accessCode, handler), accessCode, nil
}

// announceLANMode runs once the port is bound, so the QR code never points at a server that failed to start.
func announceLANMode(output io.Writer, listening net.Addr, accessCode string) error {
	_, port, err := net.SplitHostPort(listening.String())
	if err != nil {
		return fmt.Errorf("parse listening address %s: %w", listening, err)
	}
	urls, err := lanURLs(port)
	if err != nil {
		return err
	}
	if err := printLANInstructions(output, urls, accessCode); err != nil {
		return fmt.Errorf("print LAN instructions: %w", err)
	}
	return nil
}

func loadOrCreateAccessCode(path string) (string, error) {
	stored, err := os.ReadFile(path) //nolint:gosec // reason: path is derived from the operator's --data-dir flag, never from a request
	if err == nil {
		if code := strings.TrimSpace(string(stored)); code != "" {
			return code, nil
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("read LAN access code %s: %w", path, err)
	}
	secret := make([]byte, accessCodeBytes)
	if _, err := rand.Read(secret); err != nil {
		return "", fmt.Errorf("generate LAN access code: %w", err)
	}
	code := accessCodeEncoding.EncodeToString(secret)
	if err := os.WriteFile(path, []byte(code+"\n"), accessCodeFileMode); err != nil {
		return "", fmt.Errorf("store LAN access code %s: %w", path, err)
	}
	return code, nil
}

func lanListenAddress(address string) (string, error) {
	_, port, err := net.SplitHostPort(address)
	if err != nil {
		return "", fmt.Errorf("parse listen address %q: %w", address, err)
	}
	return net.JoinHostPort(allInterfaces, port), nil
}

func lanURLs(port string) ([]string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("list network interfaces: %w", err)
	}
	var urls []string
	for _, address := range addresses {
		network, ok := address.(*net.IPNet)
		if !ok || network.IP.To4() == nil || !network.IP.IsPrivate() {
			continue
		}
		urls = append(urls, "http://"+net.JoinHostPort(network.IP.String(), port)+"/")
	}
	return urls, nil
}

func printLANInstructions(output io.Writer, urls []string, accessCode string) error {
	if len(urls) == 0 {
		_, err := fmt.Fprintln(output, "LAN mode: no private IPv4 address found; is Wi-Fi connected?")
		return err
	}
	loginURL := urls[0] + "?" + url.Values{"code": {accessCode}}.Encode()
	qrCode, err := terminalqr.Render(loginURL)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "\nLAN mode: open on your phone (same Wi-Fi)\n  %s\naccess code: %s\n\n%s\n",
		strings.Join(urls, "\n  "), accessCode, qrCode)
	return err
}
