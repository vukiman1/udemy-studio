package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAccessCodeIsCreatedPrivatelyAndReused(t *testing.T) {
	path := filepath.Join(t.TempDir(), "lan-access-code")
	first, err := loadOrCreateAccessCode(path)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode().Perm() != accessCodeFileMode {
		t.Errorf("mode = %v, want %v", info.Mode().Perm(), os.FileMode(accessCodeFileMode))
	}
	second, err := loadOrCreateAccessCode(path)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if first != second || len(first) < 12 {
		t.Errorf("codes %q then %q, want the same code of at least 12 characters", first, second)
	}
}

func TestLANListenAddressKeepsThePort(t *testing.T) {
	address, err := lanListenAddress("127.0.0.1:8080")
	if err != nil || address != "0.0.0.0:8080" {
		t.Errorf("lanListenAddress = %q, %v", address, err)
	}
	if _, err := lanListenAddress("8080"); err == nil {
		t.Error("a bare port should be rejected")
	}
}

func TestLANInstructionsCarryTheCodeInTheQRLink(t *testing.T) {
	var output strings.Builder
	if err := printLANInstructions(&output, []string{"http://192.168.1.20:8080/"}, "ABCDEFGHJK"); err != nil {
		t.Fatalf("print: %v", err)
	}
	text := output.String()
	if !strings.Contains(text, "http://192.168.1.20:8080/") || !strings.Contains(text, "access code: ABCDEFGHJK") || !strings.Contains(text, "█") {
		t.Errorf("instructions = %q", text)
	}
}
