package subtitle

import (
	_ "embed"
	"encoding/json"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode/utf8"
)

const (
	hexReferenceBase        = 16
	decimalReferenceBase    = 10
	shortestEntityPrefix    = 2
	firstSurrogate          = 0xD800
	lastSurrogate           = 0xDFFF
	numericReferenceBitSize = 64
)

//go:embed html_entities.json
var htmlEntitiesJSON []byte

var characterReference = regexp.MustCompile(`&(#[0-9]+;?|#[xX][0-9a-fA-F]+;?|[^\t\n\f <&#;]{1,32};?)`)

type htmlEntityTables struct {
	named             map[string]string
	invalidCharrefs   map[uint64]string
	invalidCodepoints map[uint64]bool
}

// Python's own HTML5 tables, so references resolve exactly as html.unescape resolves them.
var htmlTables = sync.OnceValue(func() htmlEntityTables {
	var raw struct {
		HTML5             map[string]string `json:"html5"`
		InvalidCharrefs   map[string]string `json:"invalid_charrefs"`
		InvalidCodepoints []uint64          `json:"invalid_codepoints"`
	}
	if err := json.Unmarshal(htmlEntitiesJSON, &raw); err != nil {
		panic("embedded HTML entity tables are corrupt: " + err.Error())
	}
	tables := htmlEntityTables{named: raw.HTML5, invalidCharrefs: map[uint64]string{}, invalidCodepoints: map[uint64]bool{}}
	for code, text := range raw.InvalidCharrefs {
		number, err := strconv.ParseUint(code, decimalReferenceBase, numericReferenceBitSize)
		if err != nil {
			panic("embedded HTML entity tables are corrupt: " + err.Error())
		}
		tables.invalidCharrefs[number] = text
	}
	for _, code := range raw.InvalidCodepoints {
		tables.invalidCodepoints[code] = true
	}
	return tables
})

// unescapeHTML reproduces Python's html.unescape, which differs from html.UnescapeString on invalid code points.
func unescapeHTML(text string) string {
	if !strings.Contains(text, "&") {
		return text
	}
	return characterReference.ReplaceAllStringFunc(text, func(match string) string {
		reference := match[1:]
		if strings.HasPrefix(reference, "#") {
			return numericReference(strings.TrimRight(reference[1:], ";"))
		}
		return namedReference(reference)
	})
}

func numericReference(digits string) string {
	base := decimalReferenceBase
	if rest, isHex := strings.CutPrefix(strings.ToLower(digits), "x"); isHex {
		digits, base = rest, hexReferenceBase
	}
	number, err := strconv.ParseUint(digits, base, numericReferenceBitSize)
	tables := htmlTables()
	if text, ok := tables.invalidCharrefs[number]; ok && err == nil {
		return text
	}
	if errors.Is(err, strconv.ErrRange) || (number >= firstSurrogate && number <= lastSurrogate) || number > utf8.MaxRune {
		return string(utf8.RuneError)
	}
	if tables.invalidCodepoints[number] {
		return ""
	}
	return string(rune(number))
}

// A name that is not an entity resolves to its longest prefix that is one, as the HTML5 spec allows.
func namedReference(name string) string {
	tables := htmlTables()
	if text, ok := tables.named[name]; ok {
		return text
	}
	runes := []rune(name)
	for length := len(runes) - 1; length >= shortestEntityPrefix; length-- {
		if text, ok := tables.named[string(runes[:length])]; ok {
			return text + string(runes[length:])
		}
	}
	return "&" + name
}
