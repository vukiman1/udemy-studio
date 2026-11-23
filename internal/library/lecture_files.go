package library

import (
	"path"
	"regexp"
	"slices"
	"strings"
)

const (
	videoExtension   = ".mp4"
	articleExtension = ".html"
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

type lectureFiles struct {
	videos, articles, others []string
	subtitles                []subtitleCandidate
}

func (group lectureGroup) filePath(name string) string { return path.Join(group.chapterDir, name) }

func (group lectureGroup) stem(name string) string {
	return strings.TrimSuffix(strings.TrimPrefix(name, group.prefix), path.Ext(name))
}

func buildLecture(group lectureGroup, number int) Lecture {
	files := group.classify()
	lecture, others := group.mainFile(number, files)
	subtitles, unmatched := group.ownSubtitles(lecture, files.subtitles)
	lecture.Subtitles = subtitles
	lecture.Attachments = group.attachments(append(others, unmatched...))
	return lecture
}

func (group lectureGroup) classify() lectureFiles {
	var files lectureFiles
	for _, name := range group.names {
		if match := subtitleName.FindStringSubmatch(name); match != nil {
			file := SubtitleFile{Language: match[2], Path: group.filePath(name)}
			files.subtitles = append(files.subtitles, subtitleCandidate{key: matchKey(strings.TrimPrefix(match[1], group.prefix)), file: file})
			continue
		}
		switch path.Ext(name) {
		case videoExtension:
			files.videos = append(files.videos, name)
		case articleExtension:
			files.articles = append(files.articles, name)
		default:
			files.others = append(files.others, name)
		}
	}
	return files
}

// mainFile picks what the lecture plays or shows and returns every other non-subtitle file as an attachment.
func (group lectureGroup) mainFile(number int, files lectureFiles) (Lecture, []string) {
	lecture := Lecture{Number: number}
	others := files.others
	videoName, spareVideos := pickCaptionedVideo(group, files.videos, files.subtitles)
	switch {
	case len(files.articles) > 0:
		lecture.Kind, lecture.Title, lecture.ArticlePath = LectureKindArticle, group.stem(files.articles[0]), group.filePath(files.articles[0])
		others = append(others, files.articles[1:]...)
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
		lecture.Kind, lecture.Title = LectureKindUnavailable, fallbackTitle(group, files.subtitles, others)
	}
	return lecture, append(others, spareVideos...)
}

func (group lectureGroup) ownSubtitles(lecture Lecture, candidates []subtitleCandidate) ([]SubtitleFile, []string) {
	ownKeys := []string{matchKey(lecture.Title), matchKey(group.stem(path.Base(lecture.VideoPath)))}
	var own []SubtitleFile
	var unmatched []string
	for _, candidate := range candidates {
		if slices.Contains(ownKeys, candidate.key) {
			own = append(own, candidate.file)
			continue
		}
		unmatched = append(unmatched, path.Base(candidate.file.Path))
	}
	slices.SortFunc(own, func(left, right SubtitleFile) int { return strings.Compare(left.Language, right.Language) })
	return own, unmatched
}

func (group lectureGroup) attachments(names []string) []Attachment {
	slices.Sort(names)
	var attachments []Attachment
	for _, name := range names {
		attachments = append(attachments, Attachment{Name: strings.TrimPrefix(name, group.prefix), Path: group.filePath(name)})
	}
	return attachments
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
