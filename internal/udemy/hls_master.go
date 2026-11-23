package udemy

import (
	"strconv"
	"strings"
)

const (
	streamInfoTag       = "#EXT-X-STREAM-INF:"
	resolutionAttribute = "RESOLUTION"
	codecsAttribute     = "CODECS"
)

type HLSVariant struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URI    string `json:"uri"`
}

// ParseMasterPlaylist lists the variants that carry both a resolution and codecs, keeping the first variant of each height.
func ParseMasterPlaylist(playlist string) []HLSVariant {
	var variants []HLSVariant
	seenHeights := map[int]bool{}
	var pending map[string]string
	for line := range strings.Lines(playlist) {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, streamInfoTag):
			pending = parseAttributes(strings.TrimPrefix(line, streamInfoTag))
		case line == "" || strings.HasPrefix(line, "#"):
		case pending != nil:
			variant, ok := toVariant(pending, line)
			pending = nil
			if !ok || seenHeights[variant.Height] {
				continue
			}
			seenHeights[variant.Height] = true
			variants = append(variants, variant)
		}
	}
	return variants
}

func toVariant(attributes map[string]string, uri string) (HLSVariant, bool) {
	if attributes[codecsAttribute] == "" {
		return HLSVariant{}, false
	}
	widthText, heightText, found := strings.Cut(attributes[resolutionAttribute], "x")
	if !found {
		return HLSVariant{}, false
	}
	width, widthErr := strconv.Atoi(widthText)
	height, heightErr := strconv.Atoi(heightText)
	if widthErr != nil || heightErr != nil {
		return HLSVariant{}, false
	}
	return HLSVariant{Width: width, Height: height, URI: uri}, true
}

func parseAttributes(list string) map[string]string {
	attributes := map[string]string{}
	for list != "" {
		name, rest, found := strings.Cut(list, "=")
		if !found {
			break
		}
		var value string
		if strings.HasPrefix(rest, `"`) {
			value, rest, _ = strings.Cut(rest[1:], `"`)
			_, rest, _ = strings.Cut(rest, ",")
		} else {
			value, rest, _ = strings.Cut(rest, ",")
		}
		attributes[strings.TrimSpace(name)] = value
		list = rest
	}
	return attributes
}
