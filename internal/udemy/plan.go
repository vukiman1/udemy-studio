package udemy

import (
	"context"
	"fmt"
	"slices"
)

type LectureKind string

const (
	KindVideo       LectureKind = "video"
	KindHTML        LectureKind = "html"
	KindDRM         LectureKind = "drm"
	KindQuiz        LectureKind = "quiz"
	KindUnavailable LectureKind = "unavailable"
)

var quizLikeClasses = []string{"quiz", "practice", "role-play"}

type PlanOptions struct {
	Quality           int
	CaptionLanguage   string
	DownloadAssets    bool
	ContinuousNumbers bool
	SkipHLS           bool
	IDAsCourseName    bool
}

type Plan struct {
	CourseDir string        `json:"course_dir"`
	Chapters  []ChapterPlan `json:"chapters"`
	Warnings  []string      `json:"warnings,omitempty"`
}

type ChapterPlan struct {
	Index    int           `json:"index"`
	ID       int64         `json:"id"`
	Dir      string        `json:"dir"`
	Lectures []LecturePlan `json:"lectures"`
}

type LecturePlan struct {
	Number   int           `json:"number"`
	ID       int64         `json:"id"`
	Class    string        `json:"class"`
	Title    string        `json:"title"`
	Kind     LectureKind   `json:"kind"`
	File     string        `json:"file"`
	HTMLBody string        `json:"-"`
	Video    *VideoChoice  `json:"video,omitempty"`
	Captions []CaptionPlan `json:"captions,omitempty"`
	Assets   []AssetPlan   `json:"assets,omitempty"`
}

// BuildPlan decides, without downloading anything, which files a course becomes and where each one comes from.
func BuildPlan(ctx context.Context, course Course, items []CurriculumItem, options PlanOptions, playlists PlaylistFetcher) (Plan, error) {
	plan := Plan{CourseDir: courseDirName(course, options.IDAsCourseName), Chapters: []ChapterPlan{}}
	planner := lecturePlanner{
		options: options,
		sources: sourceCollector{playlists: playlists, skipHLS: options.SkipHLS, warn: plan.warn},
		warn:    plan.warn,
	}
	for _, chapter := range BuildCourseTree(items, options.ContinuousNumbers) {
		chapterPlan := ChapterPlan{Index: chapter.Index, ID: chapter.ID, Dir: chapter.Dir, Lectures: []LecturePlan{}}
		for _, lecture := range chapter.Lectures {
			lecturePlan, err := planner.plan(ctx, lecture)
			if err != nil {
				return Plan{}, err
			}
			chapterPlan.Lectures = append(chapterPlan.Lectures, lecturePlan)
		}
		plan.Chapters = append(plan.Chapters, chapterPlan)
	}
	return plan, nil
}

func (plan *Plan) warn(format string, args ...any) {
	plan.Warnings = append(plan.Warnings, fmt.Sprintf(format, args...))
}

type lecturePlanner struct {
	options PlanOptions
	sources sourceCollector
	warn    func(string, ...any)
}

func (planner lecturePlanner) plan(ctx context.Context, lecture Lecture) (LecturePlan, error) {
	plan := LecturePlan{Number: lecture.Number, ID: lecture.Item.ID, Class: lecture.Item.Class, Title: lecture.Title}
	if slices.Contains(quizLikeClasses, lecture.Item.Class) {
		plan.Kind, plan.File = KindQuiz, quizFileName(lecture.Title)
		return plan, nil
	}
	asset := lecture.Item.Asset
	if asset == nil {
		plan.Kind, plan.File = KindUnavailable, videoFileName(lecture.Title)
		planner.warn("lecture %s has no asset", lecture.Title)
		return plan, nil
	}
	if err := planner.planContent(ctx, &plan, *asset); err != nil {
		return LecturePlan{}, err
	}
	if plan.Kind == KindVideo || plan.Kind == KindDRM {
		plan.Captions = planner.planCaptions(lecture.Title, asset.Captions)
	}
	if planner.options.DownloadAssets {
		plan.Assets = assetPlanner{lecture: lecture, warn: planner.warn}.plan(*asset, lecture.Item.SupplementaryAssets)
	}
	return plan, nil
}

func (planner lecturePlanner) planContent(ctx context.Context, plan *LecturePlan, asset Asset) error {
	switch {
	case len(asset.StreamURLs) > 0:
		choices, err := planner.sources.collect(ctx, asset.StreamURLs[videoStreamKey])
		if err != nil {
			return err
		}
		plan.Kind, plan.File, plan.Video = KindVideo, videoFileName(plan.Title), SelectVideo(choices, planner.options.Quality)
	case asset.StreamURLs == nil && len(asset.MediaSources) > 0:
		plan.Kind, plan.File = KindDRM, videoFileName(plan.Title)
	default:
		plan.Kind, plan.File, plan.HTMLBody = KindHTML, htmlFileName(plan.Title), asset.Body
	}
	return nil
}

func (planner lecturePlanner) planCaptions(lectureTitle string, captions []Caption) []CaptionPlan {
	if planner.options.CaptionLanguage == "" {
		return nil
	}
	return planCaptions(lectureTitle, captions, planner.options.CaptionLanguage, planner.warn)
}
