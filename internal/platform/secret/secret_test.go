package secret

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

const tokenValue = "eyJ-very-secret-token" //nolint:gosec // reason: fake token proving it never leaks

func TestSecretNeverLeaksThroughFormattingLoggingOrJSON(t *testing.T) {
	token := New(tokenValue)
	holder := struct{ Token Secret }{Token: token}

	var logged bytes.Buffer
	slog.New(slog.NewTextHandler(&logged, nil)).Info("starting job", "token", token, "holder", holder)
	encoded, err := json.Marshal(holder)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	outputs := map[string]string{
		"%v":   fmt.Sprintf("%v", token),
		"%+v":  fmt.Sprintf("%+v", holder),
		"%#v":  fmt.Sprintf("%#v", token),
		"slog": logged.String(),
		"json": string(encoded),
	}
	for name, output := range outputs {
		if strings.Contains(output, tokenValue) {
			t.Errorf("%s leaked the token: %s", name, output)
		}
	}
}

func TestRevealReturnsTheValue(t *testing.T) {
	if New(tokenValue).Reveal() != tokenValue || !New("").IsZero() {
		t.Error("Reveal/IsZero do not reflect the wrapped value")
	}
}
