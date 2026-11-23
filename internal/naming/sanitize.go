package naming

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxFilenameBytes    = 255
	reservedSuffix      = "_"
	asciiDelete         = 0x7f
	lastASCIIControl    = 0x1f
	firstASCIISeparator = 0x1c
)

var (
	windowsReservedNames = []string{
		"CON", "PRN", "AUX", "CLOCK$", "NUL",
		"COM0", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT0", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
		"COM¹", "COM²", "COM³", "LPT¹", "LPT²", "LPT³",
		":",
	}
	leadingDots = regexp.MustCompile(`^\.{3,}`)
	rootName    = regexp.MustCompile(`^[^.]+`)
)

// SanitizeFilename reproduces pathvalidate 3.3.1 sanitize_filename so names match the Python downloader's.
func SanitizeFilename(name string) string {
	if isBlank(name) {
		return ""
	}
	sanitized := truncateUTF8(strings.Map(dropForbidden, name), maxFilenameBytes)
	if isBlank(sanitized) {
		return ""
	}
	if reserved, ok := reservedRoot(sanitized); ok {
		return strings.ReplaceAll(sanitized, reserved, reservedReplacement(reserved))
	}
	if sanitized == "." || sanitized == ".." {
		return sanitized
	}
	if strings.HasSuffix(sanitized, " ") || strings.HasSuffix(sanitized, ".") || strings.HasPrefix(sanitized, " ") {
		sanitized = strings.TrimRight(strings.TrimLeft(sanitized, " "), " ")
		if sanitized != "." && sanitized != ".." {
			sanitized = strings.TrimRight(sanitized, " .")
		}
	}
	return sanitized
}

func dropForbidden(r rune) rune {
	if r <= lastASCIIControl || r == asciiDelete || strings.ContainsRune(`/\:*?"<>|`, r) {
		return -1
	}
	return r
}

func reservedRoot(name string) (string, bool) {
	root := name
	if name != "." && name != ".." && !leadingDots.MatchString(name) {
		root = rootName.FindString(name)
	}
	for _, candidate := range []string{root, name} {
		if slices.Contains(windowsReservedNames, strings.ToUpper(candidate)) {
			return root, true
		}
	}
	return "", false
}

func reservedReplacement(reserved string) string {
	if reserved == "." || reserved == ".." {
		return reserved
	}
	return reserved + reservedSuffix
}

func truncateUTF8(text string, maxBytes int) string {
	if len(text) <= maxBytes {
		return text
	}
	truncated := text[:maxBytes]
	for !utf8.ValidString(truncated) {
		truncated = truncated[:len(truncated)-1]
	}
	return truncated
}

// Python's str.strip and str.isspace also treat the ASCII separators 0x1C-0x1F as whitespace.
func isBlank(text string) bool {
	return strings.TrimFunc(text, func(r rune) bool { return unicode.IsSpace(r) || (r >= firstASCIISeparator && r <= lastASCIIControl) }) == ""
}
