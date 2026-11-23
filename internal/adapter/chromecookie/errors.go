package chromecookie

import "errors"

var (
	ErrUnsupportedEncryption = errors.New("unsupported cookie encryption")
	ErrBadPadding            = errors.New("invalid PKCS7 padding")
	ErrHostDigestMismatch    = errors.New("cookie host digest does not match")
	ErrUnsupportedPlatform   = errors.New("reading Chrome cookies is only supported on macOS")
)
