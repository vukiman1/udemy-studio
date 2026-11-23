package subtitle

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"unicode/utf16"
	"unicode/utf8"
)

const (
	byteOrderMark  = "\ufeff"
	utf8Name       = "UTF-8"
	utf16UnitBytes = 2
	utf32UnitBytes = 4
)

var errUndecodableText = errors.New("subtitle text cannot be decoded")

type encodingMark struct {
	name   string
	prefix []byte
	decode func([]byte) (string, error)
}

// Checked in this order, as webvtt-py does, so the UTF-32 LE mark wins over the UTF-16 LE one it starts with.
var encodingMarks = []encodingMark{
	{utf8Name, []byte{0xEF, 0xBB, 0xBF}, decodeUTF8},
	{"UTF-32 LE", []byte{0xFF, 0xFE, 0x00, 0x00}, func(data []byte) (string, error) { return decodeUTF32(data, binary.LittleEndian) }},
	{"UTF-32 BE", []byte{0x00, 0x00, 0xFE, 0xFF}, func(data []byte) (string, error) { return decodeUTF32(data, binary.BigEndian) }},
	{"UTF-16 LE", []byte{0xFF, 0xFE}, func(data []byte) (string, error) { return decodeUTF16(data, binary.LittleEndian) }},
	{"UTF-16 BE", []byte{0xFE, 0xFF}, func(data []byte) (string, error) { return decodeUTF16(data, binary.BigEndian) }},
}

func decodeText(data []byte) (string, error) {
	for _, mark := range encodingMarks {
		if bytes.HasPrefix(data, mark.prefix) {
			return decodeAs(mark.name, mark.decode, data[len(mark.prefix):])
		}
	}
	return decodeAs(utf8Name, decodeUTF8, data)
}

func decodeAs(name string, decode func([]byte) (string, error), data []byte) (string, error) {
	text, err := decode(data)
	if err != nil {
		return "", fmt.Errorf("decode %d bytes as %s: %w", len(data), name, err)
	}
	return text, nil
}

func decodeUTF8(data []byte) (string, error) {
	if !utf8.Valid(data) {
		return "", errUndecodableText
	}
	return string(data), nil
}

// Unlike utf16.Decode, an unpaired surrogate fails the decoding instead of becoming U+FFFD, as in Python.
func decodeUTF16(data []byte, order binary.ByteOrder) (string, error) {
	if len(data)%utf16UnitBytes != 0 {
		return "", errUndecodableText
	}
	var runes []rune
	for offset := 0; offset < len(data); offset += utf16UnitBytes {
		unit := rune(order.Uint16(data[offset:]))
		if !utf16.IsSurrogate(unit) {
			runes = append(runes, unit)
			continue
		}
		next := offset + utf16UnitBytes
		if next >= len(data) {
			return "", errUndecodableText
		}
		paired := utf16.DecodeRune(unit, rune(order.Uint16(data[next:])))
		if paired == utf8.RuneError {
			return "", errUndecodableText
		}
		runes = append(runes, paired)
		offset = next
	}
	return string(runes), nil
}

func decodeUTF32(data []byte, order binary.ByteOrder) (string, error) {
	if len(data)%utf32UnitBytes != 0 {
		return "", errUndecodableText
	}
	runes := make([]rune, 0, len(data)/utf32UnitBytes)
	for offset := 0; offset < len(data); offset += utf32UnitBytes {
		value := order.Uint32(data[offset:])
		if value > utf8.MaxRune || !utf8.ValidRune(rune(value)) {
			return "", errUndecodableText
		}
		runes = append(runes, rune(value))
	}
	return string(runes), nil
}
