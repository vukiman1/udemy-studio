package cli

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"github.com/vukiman1/udemy-studio/internal/app"
	"github.com/vukiman1/udemy-studio/internal/udemy"
)

func newPlanCommand() *cobra.Command {
	var options app.PlanCourseOptions
	command := &cobra.Command{
		Use:   "plan <course-url>",
		Short: "Print, as JSON, what downloading a course would fetch and where each file would go",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			options.CourseURL = args[0]
			if err := applyDataDefaults(&options.DataDir, &options.ChromeDataDir); err != nil {
				return err
			}
			logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
			plan, err := app.PlanCourse(command.Context(), options, logger)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(command.OutOrStdout())
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(plan); err != nil {
				return fmt.Errorf("print plan: %w", err)
			}
			return nil
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.DataDir, "data-dir", "", dataDirUsage)
	flags.StringVar(&options.ChromeDataDir, "chrome-data-dir", "", chromeDataDirUsage)
	flags.IntVarP(&options.Plan.Quality, "quality", "q", udemy.BestQuality, "video height to aim for; the nearest one wins (default: highest)")
	flags.StringVarP(&options.Plan.CaptionLanguage, "caption-language", "l", defaultCaptionLang, `caption language, "all", or "" for none`)
	flags.BoolVar(&options.Plan.DownloadAssets, "assets", true, "plan attachments, source code and external links")
	flags.BoolVarP(&options.Plan.ContinuousNumbers, "continuous-numbers", "n", false, "number lectures across the whole course instead of per chapter")
	flags.BoolVar(&options.Plan.SkipHLS, "skip-hls", false, "ignore HLS streams and use only direct MP4 sources")
	flags.BoolVar(&options.Plan.IDAsCourseName, "id-as-course-name", false, "name the course folder by its numeric id")
	return command
}
