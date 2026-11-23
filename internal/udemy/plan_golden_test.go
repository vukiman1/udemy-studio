package udemy

import (
	"encoding/json"
	"io/fs"
	"maps"
	"slices"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/vukiman1/udemy-studio/internal/naming"
)

const pythonBestQuality = "None"

type curriculumFixture struct {
	Results    []CurriculumItem `json:"results"`
	CourseID   int64            `json:"course_id"`
	CourseSlug string           `json:"course_slug"`
}

type goldenCoursePlan struct {
	ContinuousNumbers bool            `json:"continuous_numbers"`
	SkipHLS           bool            `json:"skip_hls"`
	CourseDir         string          `json:"course_dir"`
	Chapters          []goldenChapter `json:"chapters"`
}

type goldenChapter struct {
	Index    int             `json:"index"`
	ID       int64           `json:"id"`
	Dir      string          `json:"dir"`
	Lectures []goldenLecture `json:"lectures"`
}

type goldenLecture struct {
	Index    int                     `json:"index"`
	ID       int64                   `json:"id"`
	Class    string                  `json:"class"`
	Title    string                  `json:"title"`
	Kind     LectureKind             `json:"kind"`
	File     string                  `json:"file"`
	HasBody  bool                    `json:"has_body"`
	Choices  map[string]*VideoChoice `json:"choices"`
	Captions []CaptionPlan           `json:"captions"`
	Assets   []AssetPlan             `json:"assets"`
}

type comparableLecture struct {
	Number   int
	ID       int64
	Class    string
	Title    string
	Kind     LectureKind
	File     string
	HasBody  bool
	Video    *VideoChoice
	Captions []CaptionPlan
	Assets   []AssetPlan
}

func TestPlanMatchesPythonDownloader(t *testing.T) {
	fixtures := []struct{ curriculum, golden string }{
		{"udemy/python/course_content.json", "udemy/plan/plan_golden.json"},
		{"udemy/python/edge_course_content.json", "udemy/plan/edge_plan_golden.json"},
	}
	for _, fixture := range fixtures {
		curriculum := readJSON[curriculumFixture](t, fixture.curriculum)
		course := Course{ID: curriculum.CourseID, Slug: curriculum.CourseSlug}
		for _, golden := range readJSON[[]goldenCoursePlan](t, fixture.golden) {
			for _, quality := range goldenQualities(golden) {
				name := fixture.golden + "/continuous=" + strconv.FormatBool(golden.ContinuousNumbers) + "/quality=" + quality
				t.Run(name, func(t *testing.T) {
					options := PlanOptions{
						Quality:           pythonQuality(t, quality),
						CaptionLanguage:   AllCaptionLanguages,
						DownloadAssets:    true,
						ContinuousNumbers: golden.ContinuousNumbers,
						SkipHLS:           golden.SkipHLS,
					}
					plan, err := BuildPlan(t.Context(), course, curriculum.Results, options, fixturePlaylists{})
					if err != nil {
						t.Fatalf("BuildPlan: %v", err)
					}
					compareWithGolden(t, plan, golden, quality)
				})
			}
		}
	}
}

func compareWithGolden(t *testing.T, plan Plan, golden goldenCoursePlan, quality string) {
	t.Helper()
	if plan.CourseDir != golden.CourseDir {
		t.Errorf("course dir = %q, python %q", plan.CourseDir, golden.CourseDir)
	}
	if len(plan.Chapters) != len(golden.Chapters) {
		t.Fatalf("got %d chapters, python %d", len(plan.Chapters), len(golden.Chapters))
	}
	for chapterIndex, chapter := range plan.Chapters {
		want := golden.Chapters[chapterIndex]
		if chapter.Index != want.Index || chapter.ID != want.ID || chapter.Dir != want.Dir {
			t.Errorf("chapter %d = (%d, %d, %q), python (%d, %d, %q)", chapterIndex, chapter.Index, chapter.ID, chapter.Dir, want.Index, want.ID, want.Dir)
		}
		got := make([]comparableLecture, len(chapter.Lectures))
		for index, lecture := range chapter.Lectures {
			got[index] = comparableFromPlan(lecture)
		}
		expected := make([]comparableLecture, len(want.Lectures))
		for index, lecture := range want.Lectures {
			expected[index] = comparableFromGolden(lecture, quality)
		}
		if diff := cmp.Diff(expected, got); diff != "" {
			t.Errorf("chapter %q lectures differ from python (-python +go):\n%s", want.Dir, diff)
		}
	}
}

func comparableFromPlan(lecture LecturePlan) comparableLecture {
	return comparableLecture{
		Number: lecture.Number, ID: lecture.ID, Class: lecture.Class, Title: lecture.Title, Kind: lecture.Kind,
		File: lecture.File, HasBody: lecture.HTMLBody != "", Video: lecture.Video, Captions: lecture.Captions, Assets: lecture.Assets,
	}
}

func comparableFromGolden(lecture goldenLecture, quality string) comparableLecture {
	assets := make([]AssetPlan, 0, len(lecture.Assets))
	for _, asset := range lecture.Assets {
		// intended difference: Go sanitizes asset file names, Python joins Udemy's raw filename into the path
		asset.File = naming.SanitizeFilename(asset.File)
		assets = append(assets, asset)
	}
	return comparableLecture{
		Number: lecture.Index, ID: lecture.ID, Class: lecture.Class, Title: lecture.Title, Kind: lecture.Kind,
		File: lecture.File, HasBody: lecture.HasBody, Video: lecture.Choices[quality], Captions: emptyAsNil(lecture.Captions), Assets: emptyAsNil(assets),
	}
}

func goldenQualities(golden goldenCoursePlan) []string {
	for _, chapter := range golden.Chapters {
		for _, lecture := range chapter.Lectures {
			if lecture.Choices != nil {
				return slices.Collect(maps.Keys(lecture.Choices))
			}
		}
	}
	return []string{pythonBestQuality}
}

func pythonQuality(t *testing.T, quality string) int {
	t.Helper()
	if quality == pythonBestQuality {
		return BestQuality
	}
	value, err := strconv.Atoi(quality)
	if err != nil {
		t.Fatalf("golden quality %q: %v", quality, err)
	}
	return value
}

func emptyAsNil[T any](items []T) []T {
	if len(items) == 0 {
		return nil
	}
	return items
}

func readJSON[T any](t *testing.T, path string) T {
	t.Helper()
	raw, err := fs.ReadFile(testdata, path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return value
}
