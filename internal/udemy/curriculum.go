package udemy

type Course struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Slug   string `json:"published_title"`
	Portal string `json:"-"`
}

type CurriculumItem struct {
	Class               string  `json:"_class"`
	ID                  int64   `json:"id"`
	ObjectIndex         int     `json:"object_index"`
	Title               string  `json:"title"`
	Asset               *Asset  `json:"asset"`
	SupplementaryAssets []Asset `json:"supplementary_assets"`
}

type Asset struct {
	ID          int64  `json:"id"`
	AssetType   string `json:"asset_type"`
	Filename    string `json:"filename"`
	Body        string `json:"body"`
	ExternalURL string `json:"external_url"`
	// nil means Udemy sent no stream_urls (DRM or HTML), an empty non-nil map means an HTML lecture.
	StreamURLs   map[string][]VideoSource `json:"stream_urls"`
	MediaSources []MediaSource            `json:"media_sources"`
	Captions     []Caption                `json:"captions"`
	DownloadURLs map[string][]DownloadURL `json:"download_urls"`
}

type VideoSource struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	File  string `json:"file"`
}

type MediaSource struct {
	Type string `json:"type"`
	Src  string `json:"src"`
}

type Caption struct {
	Class    string `json:"_class"`
	URL      string `json:"url"`
	Language string `json:"language"`
	Srclang  string `json:"srclang"`
	Label    string `json:"label"`
	LocaleID string `json:"locale_id"`
}

type DownloadURL struct {
	File string `json:"file"`
}
