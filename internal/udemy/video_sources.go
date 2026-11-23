package udemy

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type SourceType string

const (
	SourceMP4 SourceType = "video"
	SourceHLS SourceType = "hls"
)

const (
	videoStreamKey   = "Video"
	audioLabel       = "audio"
	hlsMimeType      = "application/x-mpegURL"
	hlsExtension     = "m3u8"
	unknownWidth     = 256
	hlsPlaylistFile  = "index_%s_%dx%d.m3u8"
	masterAssetGroup = 1
)

// BestQuality asks for the tallest video instead of the one nearest a height.
const BestQuality = 0

var (
	widthForHeight = map[int]int{2160: 3840, 1440: 2560, 1080: 1920, 720: 1280, 480: 854, 360: 640, 240: 426}
	masterAssetID  = regexp.MustCompile(`assets/(\d+)/`)
)

// ErrPlaylistUnavailable marks a playlist Udemy refused or no longer has, which only costs that lecture its HLS streams.
var ErrPlaylistUnavailable = errors.New("playlist unavailable")

type PlaylistFetcher interface {
	FetchPlaylist(ctx context.Context, playlistURL string) (string, error)
}

type VideoChoice struct {
	Type         SourceType `json:"type"`
	Height       int        `json:"height"`
	Width        int        `json:"width"`
	URL          string     `json:"url"`
	PlaylistFile string     `json:"playlist_file,omitempty"`
}

type sourceCollector struct {
	playlists PlaylistFetcher
	skipHLS   bool
	warn      func(format string, args ...any)
}

func (collector sourceCollector) collect(ctx context.Context, sources []VideoSource) ([]VideoChoice, error) {
	var choices []VideoChoice
	for _, source := range sources {
		if source.File == "" || strings.EqualFold(source.Label, audioLabel) {
			continue
		}
		if source.Type == hlsMimeType || strings.Contains(source.File, hlsExtension) {
			if collector.skipHLS {
				continue
			}
			variants, err := collector.hlsChoices(ctx, source.File)
			if err != nil {
				return nil, err
			}
			choices = append(choices, variants...)
			continue
		}
		height, err := strconv.Atoi(source.Label)
		if err != nil {
			collector.warn("skipping video source with non-numeric label %q", source.Label)
			continue
		}
		choices = append(choices, VideoChoice{Type: SourceMP4, Height: height, Width: widthOf(height), URL: source.File})
	}
	return choices, nil
}

func (collector sourceCollector) hlsChoices(ctx context.Context, masterURL string) ([]VideoChoice, error) {
	assetID := masterAssetID.FindStringSubmatch(masterURL)
	if assetID == nil {
		collector.warn("skipping HLS source without an asset id in its URL")
		return nil, nil
	}
	playlist, err := collector.playlists.FetchPlaylist(ctx, masterURL)
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if errors.Is(err, ErrPlaylistUnavailable) {
		collector.warn("skipping HLS streams of asset %s: %v", assetID[masterAssetGroup], err)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("fetch HLS master of asset %s: %w", assetID[masterAssetGroup], err)
	}
	var choices []VideoChoice
	for _, variant := range ParseMasterPlaylist(playlist) {
		choices = append(choices, VideoChoice{
			Type:         SourceHLS,
			Height:       variant.Height,
			Width:        variant.Width,
			URL:          resolveReference(masterURL, variant.URI),
			PlaylistFile: fmt.Sprintf(hlsPlaylistFile, assetID[masterAssetGroup], variant.Width, variant.Height),
		})
	}
	return choices, nil
}

// SelectVideo picks the tallest source, or with a quality the one whose height is nearest (ties go to the taller one).
func SelectVideo(choices []VideoChoice, quality int) *VideoChoice {
	if len(choices) == 0 {
		return nil
	}
	ranked := slices.Clone(choices)
	slices.SortStableFunc(ranked, func(left, right VideoChoice) int { return cmp.Compare(right.Height, left.Height) })
	if quality == BestQuality {
		return &ranked[0]
	}
	best := ranked[0]
	for _, choice := range ranked[1:] {
		if distance(choice.Height, quality) < distance(best.Height, quality) {
			best = choice
		}
	}
	return &best
}

func widthOf(height int) int {
	if width, ok := widthForHeight[height]; ok {
		return width
	}
	return unknownWidth
}

func distance(height, quality int) int {
	if height > quality {
		return height - quality
	}
	return quality - height
}

func resolveReference(base, reference string) string {
	referenceURL, err := url.Parse(reference)
	if err != nil || referenceURL.IsAbs() {
		return reference
	}
	baseURL, err := url.Parse(base)
	if err != nil {
		return reference
	}
	return baseURL.ResolveReference(referenceURL).String()
}
