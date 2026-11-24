package pyjson

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

const (
	itemSeparator     = ", "
	keySeparator      = ": "
	firstPrintable    = ' '
	lastPrintable     = '~'
	unicodeEscape     = `\u%04x`
	pythonInfinity    = "Infinity"
	pythonNegInfinity = "-Infinity"
	maxFixedDecimal   = 16
	minFixedDecimal   = -4
	minExponentDigits = 2
	lastBMPCharacter  = 0xFFFF
)

var errUnsupportedType = errors.New("pyjson cannot encode this type")

var (
	integerLiteral = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
	shortEscapes   = map[rune]string{'"': `\"`, '\\': `\\`, '\n': `\n`, '\r': `\r`, '\t': `\t`, '\b': `\b`, '\f': `\f`}
)

// Dumps formats value like Python's json.dumps with its defaults: ", " and ": " separators and every non-ASCII character escaped.
func Dumps(value Value) (string, error) {
	var output strings.Builder
	if err := write(&output, value); err != nil {
		return "", err
	}
	return output.String(), nil
}

func write(output *strings.Builder, value Value) error {
	switch value := value.(type) {
	case nil:
		output.WriteString("null")
	case bool:
		output.WriteString(strconv.FormatBool(value))
	case int64:
		output.WriteString(strconv.FormatInt(value, 10))
	case Number:
		output.WriteString(formatNumber(value))
	case string:
		writeString(output, value)
	case []Value:
		return writeArray(output, value)
	case *Object:
		return writeObject(output, value)
	default:
		return fmt.Errorf("%w: %T", errUnsupportedType, value)
	}
	return nil
}

func writeArray(output *strings.Builder, items []Value) error {
	output.WriteByte('[')
	for position, item := range items {
		if position > 0 {
			output.WriteString(itemSeparator)
		}
		if err := write(output, item); err != nil {
			return err
		}
	}
	output.WriteByte(']')
	return nil
}

func writeObject(output *strings.Builder, object *Object) error {
	output.WriteByte('{')
	for position, key := range object.keys {
		if position > 0 {
			output.WriteString(itemSeparator)
		}
		writeString(output, key)
		output.WriteString(keySeparator)
		if err := write(output, object.values[key]); err != nil {
			return err
		}
	}
	output.WriteByte('}')
	return nil
}

func writeString(output *strings.Builder, text string) {
	output.WriteByte('"')
	for _, character := range text {
		if escape, ok := shortEscapes[character]; ok {
			output.WriteString(escape)
			continue
		}
		if character >= firstPrintable && character <= lastPrintable {
			output.WriteRune(character)
			continue
		}
		if character > lastBMPCharacter {
			high, low := utf16.EncodeRune(character)
			fmt.Fprintf(output, unicodeEscape+unicodeEscape, high, low)
			continue
		}
		fmt.Fprintf(output, unicodeEscape, character)
	}
	output.WriteByte('"')
}

func formatNumber(literal Number) string {
	if integerLiteral.MatchString(string(literal)) {
		integer, _ := new(big.Int).SetString(string(literal), 10)
		return integer.String()
	}
	float, err := strconv.ParseFloat(string(literal), 64)
	switch {
	case err != nil && math.IsInf(float, 1):
		return pythonInfinity
	case err != nil && math.IsInf(float, -1):
		return pythonNegInfinity
	}
	return floatRepr(float)
}

// floatRepr is Python's repr of a float: the shortest round-trip digits, in fixed notation unless the exponent is
// below -4 or above 15, with ".0" kept on whole numbers.
func floatRepr(float float64) string {
	scientific := strconv.FormatFloat(float, 'e', -1, 64)
	mantissa, exponentText, _ := strings.Cut(scientific, "e")
	exponent, _ := strconv.Atoi(exponentText)
	sign := ""
	if strings.HasPrefix(mantissa, "-") {
		sign, mantissa = "-", mantissa[1:]
	}
	digits := strings.Replace(mantissa, ".", "", 1)
	decimalPoint := exponent + 1
	if decimalPoint > maxFixedDecimal || decimalPoint <= minFixedDecimal {
		return sign + mantissa + "e" + exponentSign(exponent) + fmt.Sprintf("%0*d", minExponentDigits, abs(exponent))
	}
	switch {
	case decimalPoint <= 0:
		return sign + "0." + strings.Repeat("0", -decimalPoint) + digits
	case decimalPoint >= len(digits):
		return sign + digits + strings.Repeat("0", decimalPoint-len(digits)) + ".0"
	default:
		return sign + digits[:decimalPoint] + "." + digits[decimalPoint:]
	}
}

func exponentSign(exponent int) string {
	if exponent < 0 {
		return "-"
	}
	return "+"
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
