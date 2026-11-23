package udemy

import "strings"

const (
	captionClass        = "caption"
	vttURLMarker        = "vtt"
	AllCaptionLanguages = "all"
)

type CaptionPlan struct {
	Language string `json:"language"`
	URL      string `json:"url"`
	File     string `json:"file"`
	SRTFile  string `json:"srt_file"`
}

func planCaptions(lectureTitle string, captions []Caption, wantedLanguage string, warn func(string, ...any)) []CaptionPlan {
	var plans []CaptionPlan
	for _, caption := range captions {
		if caption.Class != captionClass || caption.URL == "" {
			continue
		}
		language := captionLanguage(caption)
		if language == "" {
			warn("skipping caption without a language: %s", lectureTitle)
			continue
		}
		if wantedLanguage != AllCaptionLanguages && language != wantedLanguage {
			continue
		}
		stem := captionStem(lectureTitle, language)
		plans = append(plans, CaptionPlan{
			Language: language,
			URL:      caption.URL,
			File:     stem + captionExtension(caption.URL),
			SRTFile:  stem + srtExtension,
		})
	}
	return plans
}

func captionLanguage(caption Caption) string {
	for _, candidate := range []string{caption.Language, caption.Srclang, caption.Label} {
		if candidate != "" {
			return candidate
		}
	}
	language, _, _ := strings.Cut(caption.LocaleID, "_")
	return language
}

func captionExtension(captionURL string) string {
	lastPart := captionURL[strings.LastIndex(captionURL, ".")+1:]
	if strings.Contains(lastPart, vttURLMarker) {
		return vttExtension
	}
	return srtExtension
}
