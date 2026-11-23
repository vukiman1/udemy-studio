package secret

import "log/slog"

const redacted = "[REDACTED]"

type Secret struct {
	value string
}

func New(value string) Secret { return Secret{value: value} }

func (secret Secret) Reveal() string { return secret.value }

func (secret Secret) IsZero() bool { return secret.value == "" }

func (Secret) String() string { return redacted }

func (Secret) GoString() string { return redacted }

func (Secret) LogValue() slog.Value { return slog.StringValue(redacted) }

func (Secret) MarshalText() ([]byte, error) { return []byte(redacted), nil }
