package chromecookie

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/vukiman1/udemy-studio/internal/login"
)

var (
	ErrUnsupportedEncryption = errors.New("unsupported cookie encryption")
	ErrBadPadding            = errors.New("invalid PKCS7 padding")
	ErrHostDigestMismatch    = errors.New("cookie host digest does not match")
	ErrUnsupportedPlatform   = errors.New("reading Chrome cookies is only supported on macOS")
)

func wrapAccessError(err error) error {
	if errors.Is(err, fs.ErrPermission) {
		return fmt.Errorf("%w (grant Full Disk Access to the app running udemy-studio): %w", login.ErrNoDiskAccess, err)
	}
	return err
}
