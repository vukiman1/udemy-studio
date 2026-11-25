package library

type LectureKind string

const (
	LectureKindVideo       LectureKind = "video"
	LectureKindArticle     LectureKind = "article"
	LectureKindUnavailable LectureKind = "unavailable"
	// LectureKindQuiz is a page that needs its own scripts to show the questions.
	LectureKindQuiz LectureKind = "quiz"
)

// CoverFileName is the course's cover image, kept beside its chapter folders.
const CoverFileName = "cover.jpg"

type Course struct {
	ID    string
	Title string
	// Dir is the course's folder in the library, holding everything downloaded with it.
	Dir string
	// CoverPath is empty when the course was downloaded without a cover.
	CoverPath string
	Chapters  []Chapter
}

type Chapter struct {
	Index    int
	Title    string
	Lectures []Lecture
}

type Lecture struct {
	ID  string
	Key string
	// PathKey is the key progress was saved under before a manifest gave the lecture its Udemy id.
	PathKey     string
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
