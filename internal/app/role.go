package app

import (
	"errors"
	"fmt"
)

type Role string

const (
	RoleLibrary Role = "library"
	RoleCloud   Role = "cloud"
)

var ErrUnknownRole = errors.New("unknown role; use library or cloud")

func ParseRole(raw string) (Role, error) {
	switch role := Role(raw); role {
	case RoleLibrary, RoleCloud:
		return role, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownRole, raw)
	}
}
