package library_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"testing"
	"testing/fstest"

	"github.com/vukiman1/udemy-studio/internal/library"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

const continuousNumbersPastThreeDigits = 1000

type curriculumFixture struct {
	Results    []udemy.CurriculumItem `json:"results"`
	CourseID   int64                  `json:"course_id"`
	CourseSlug string                 `json:"course_slug"`
}

func TestScannerFindsEveryFileThePlannerNames(t *testing.T) {
	for _, fixture := range []string{"course_content.json", "edge_course_content.json"} {
		t.Run(fixture, func(t *testing.T) {
			plan := planFixture(t, fixture)
			scanned, err := library.Scan(downloadsFromPlan(plan))
			if err != nil {
				t.Fatalf("Scan: %v", err)
			}
			found := scannedPaths(scanned)
			for _, planned := range plannedPaths(plan) {
				if !found[planned] {
					t.Errorf("planned file %q is missing from the library", planned)
				}
			}
		})
	}
}

func planFixture(t *testing.T, fixture string) udemy.Plan {
	t.Helper()
	raw, err := fs.ReadFile(os.DirFS("../../testdata/udemy/python"), fixture)
	if err != nil {
		t.Fatalf("read %s: %v", fixture, err)
	}
	var content curriculumFixture
	if err := json.Unmarshal(raw, &content); err != nil {
		t.Fatalf("decode %s: %v", fixture, err)
	}
	items := content.Results
	for range continuousNumbersPastThreeDigits {
		items = append(items, udemy.CurriculumItem{Class: "lecture", ID: 1, Title: "Filler", Asset: &udemy.Asset{AssetType: "Article"}})
	}
	options := udemy.PlanOptions{CaptionLanguage: udemy.AllCaptionLanguages, DownloadAssets: true, ContinuousNumbers: true, SkipHLS: true}
	plan, err := udemy.BuildPlan(t.Context(), udemy.Course{ID: content.CourseID, Slug: content.CourseSlug}, items, options, nil)
	if err != nil {
		t.Fatalf("BuildPlan: %v", err)
	}
	return plan
}

func downloadsFromPlan(plan udemy.Plan) fstest.MapFS {
	downloads := fstest.MapFS{}
	for _, planned := range plannedPaths(plan) {
		downloads[planned] = &fstest.MapFile{}
	}
	return downloads
}

func plannedPaths(plan udemy.Plan) []string {
	var paths []string
	for _, chapter := range plan.Chapters {
		dir := path.Join(plan.CourseDir, chapter.Dir)
		for _, lecture := range chapter.Lectures {
			if lecture.Kind == udemy.KindUnavailable || (lecture.Kind == udemy.KindVideo && lecture.Video == nil) {
				continue
			}
			paths = append(paths, path.Join(dir, lecture.File))
			for _, caption := range lecture.Captions {
				paths = append(paths, path.Join(dir, caption.SRTFile))
			}
			for _, asset := range lecture.Assets {
				paths = append(paths, path.Join(dir, asset.File))
			}
		}
	}
	return paths
}

func scannedPaths(scanned *library.Library) map[string]bool {
	found := map[string]bool{}
	for _, course := range scanned.Courses {
		for _, chapter := range course.Chapters {
			for _, lecture := range chapter.Lectures {
				found[lecture.VideoPath], found[lecture.ArticlePath] = true, true
				for _, subtitle := range lecture.Subtitles {
					found[subtitle.Path] = true
				}
				for _, attachment := range lecture.Attachments {
					found[attachment.Path] = true
				}
			}
		}
	}
	return found
}
