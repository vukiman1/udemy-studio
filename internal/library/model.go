package library

type LectureKind string

const (
	LectureKindVideo       LectureKind = "video"
	LectureKindArticle     LectureKind = "article"
	LectureKindUnavailable LectureKind = "unavailable"
)

type Course struct {
	ID       string
	Title    string
	Chapters []Chapter
}

type Chapter struct {
	Index    int
	Title    string
	Lectures []Lecture
}

type Lecture struct {
	ID          string
	Key         string
	Number      int
	Title       string
	Kind        LectureKind
	VideoPath   string
	ArticlePath string
	Subtitles   []SubtitleFile
	Attachments []Attachment
}

type SubtitleFile struct {
	Language string
	Path     string
}

type Attachment struct {
	ID   string
	Name string
	Path string
}
