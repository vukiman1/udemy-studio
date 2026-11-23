package library

import (
	"path"
	"regexp"
	"slices"
	"strings"
)

const (
	videoExtension    = ".mp4"
	articleExtension  = ".html"
	subtitleExtension = ".srt"
)

var subtitleName = regexp.MustCompile(`^(.+)_([a-z]{2,3}(?:[-_][A-Za-z]{2,4})?)\.srt$`)

type lectureGroup struct {
	chapterDir string
	prefix     string
	names      []string
}

type subtitleCandidate struct {
	key  string
	file SubtitleFile
}

func (group lectureGroup) filePath(name string) string { return path.Join(group.chapterDir, name) }

func (group lectureGroup) stem(name string) string {
	return strings.TrimSuffix(strings.TrimPrefix(name, group.prefix), path.Ext(name))
}

func buildLecture(group lectureGroup, number int) Lecture {
	var videos, articles, others []string
	var subtitles []subtitleCandidate
	for _, name := range group.names {
		if match := subtitleName.FindStringSubmatch(name); match != nil {
			file := SubtitleFile{Language: match[2], Path: group.filePath(name)}
			subtitles = append(subtitles, subtitleCandidate{key: matchKey(strings.TrimPrefix(match[1], group.prefix)), file: file})
			continue
		}
		switch path.Ext(name) {
		case videoExtension:
			videos = append(videos, name)
		case articleExtension:
			articles = append(articles, name)
		default:
			others = append(others, name)
		}
	}

	lecture := Lecture{Number: number}
	videoName, spareVideos := pickCaptionedVideo(group, videos, subtitles)
	switch {
	case len(articles) > 0:
		lecture.Kind, lecture.Title, lecture.ArticlePath = LectureKindArticle, group.stem(articles[0]), group.filePath(articles[0])
		others = append(others, articles[1:]...)
		if videoName == "" && len(spareVideos) == 1 {
			videoName, spareVideos = spareVideos[0], nil
		}
		if videoName != "" {
			lecture.VideoPath = group.filePath(videoName)
		}
	case videoName != "":
		lecture.Kind, lecture.Title, lecture.VideoPath = LectureKindVideo, group.stem(videoName), group.filePath(videoName)
	case len(spareVideos) > 0:
		lecture.Kind, lecture.Title, lecture.VideoPath = LectureKindVideo, group.stem(spareVideos[0]), group.filePath(spareVideos[0])
		spareVideos = spareVideos[1:]
	default:
		lecture.Kind, lecture.Title = LectureKindUnavailable, fallbackTitle(group, subtitles, others)
	}
	others = append(others, spareVideos...)

	ownKeys := []string{matchKey(lecture.Title), matchKey(group.stem(path.Base(lecture.VideoPath)))}
	for _, candidate := range subtitles {
		if slices.Contains(ownKeys, candidate.key) {
			lecture.Subtitles = append(lecture.Subtitles, candidate.file)
			continue
		}
		others = append(others, path.Base(candidate.file.Path))
	}
	slices.SortFunc(lecture.Subtitles, func(left, right SubtitleFile) int { return strings.Compare(left.Language, right.Language) })
	slices.Sort(others)
	for _, name := range others {
		lecture.Attachments = append(lecture.Attachments, Attachment{Name: strings.TrimPrefix(name, group.prefix), Path: group.filePath(name)})
	}
	return lecture
}

func pickCaptionedVideo(group lectureGroup, videos []string, subtitles []subtitleCandidate) (string, []string) {
	for position, video := range videos {
		key := matchKey(group.stem(video))
		if slices.ContainsFunc(subtitles, func(candidate subtitleCandidate) bool { return candidate.key == key }) {
			return video, slices.Delete(slices.Clone(videos), position, position+1)
		}
	}
	return "", videos
}

func fallbackTitle(group lectureGroup, subtitles []subtitleCandidate, others []string) string {
	if len(subtitles) > 0 {
		base := subtitleName.FindStringSubmatch(path.Base(subtitles[0].file.Path))[1]
		return strings.TrimPrefix(base, group.prefix)
	}
	return group.stem(others[0])
}
