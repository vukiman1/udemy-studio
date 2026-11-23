package udemy

import "strings"

type AssetKind string

const (
	AssetArticle      AssetKind = "article"
	AssetEbook        AssetKind = "ebook"
	AssetFile         AssetKind = "file"
	AssetPresentation AssetKind = "presentation"
	AssetAudio        AssetKind = "audio"
	AssetSourceCode   AssetKind = "source_code"
	AssetExternalLink AssetKind = "external_link"
)

const (
	articleAssetType      = "article"
	videoAssetType        = "video"
	externalLinkAssetType = "externallink"
)

type downloadableAsset struct {
	urlKey string
	kind   AssetKind
}

var (
	lectureDownloads = map[string]downloadableAsset{
		"e-book":       {"E-Book", AssetEbook},
		"file":         {"File", AssetFile},
		"presentation": {"Presentation", AssetPresentation},
		"audio":        {"Audio", AssetAudio},
	}
	supplementaryDownloads = map[string]downloadableAsset{
		"file":       {"File", AssetFile},
		"sourcecode": {"SourceCode", AssetSourceCode},
	}
)

type AssetPlan struct {
	Type AssetKind `json:"type"`
	File string    `json:"file"`
	URL  string    `json:"url,omitempty"`
}

type assetPlanner struct {
	lecture Lecture
	warn    func(string, ...any)
}

func (planner assetPlanner) plan(asset Asset, supplementary []Asset) []AssetPlan {
	var plans []AssetPlan
	assetType := strings.ToLower(asset.AssetType)
	switch assetType {
	case articleAssetType:
		plans = append(plans, AssetPlan{Type: AssetArticle, File: htmlFileName(planner.lecture.Title)})
	case videoAssetType:
		// The lecture's own video is planned with the lecture, not as an asset.
	default:
		if download, ok := lectureDownloads[assetType]; ok {
			plans = planner.appendDownload(plans, asset, download)
		} else {
			planner.warn("unknown asset type %q in %s", asset.AssetType, planner.lecture.Title)
		}
	}
	for _, entry := range supplementary {
		plans = planner.appendSupplementary(plans, entry)
	}
	return plans
}

func (planner assetPlanner) appendSupplementary(plans []AssetPlan, entry Asset) []AssetPlan {
	entryType := strings.ToLower(entry.AssetType)
	if entryType == externalLinkAssetType {
		return append(plans, AssetPlan{
			Type: AssetExternalLink,
			File: assetFileName(planner.lecture.Number, entry.Filename+urlExtension),
			URL:  entry.ExternalURL,
		})
	}
	download, ok := supplementaryDownloads[entryType]
	if !ok {
		planner.warn("skipping supplementary asset of type %q in %s", entry.AssetType, planner.lecture.Title)
		return plans
	}
	return planner.appendDownload(plans, entry, download)
}

func (planner assetPlanner) appendDownload(plans []AssetPlan, asset Asset, download downloadableAsset) []AssetPlan {
	if len(asset.DownloadURLs) == 0 {
		return plans
	}
	urls := asset.DownloadURLs[download.urlKey]
	if len(urls) == 0 {
		planner.warn("asset %q in %s has no %s download url", asset.Filename, planner.lecture.Title, download.urlKey)
		return plans
	}
	return append(plans, AssetPlan{Type: download.kind, File: assetFileName(planner.lecture.Number, asset.Filename), URL: urls[0].File})
}
