//go:build !darwin

package chromecookie

import "context"

type Keychain struct{}

func (Keychain) SafeStoragePassword(context.Context) ([]byte, error) {
	return nil, ErrUnsupportedPlatform
}
