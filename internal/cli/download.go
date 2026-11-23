package cli

import (
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/vukiman1/udemy-studio/internal/app"
	"github.com/vukiman1/udemy-studio/internal/jobs"
)

const outputFlag = "output"

func newDownloadCommand() *cobra.Command {
	var options app.DownloadCourseOptions
	command := &cobra.Command{
		Use:   "download <course-url>",
		Short: "Download a course with the Go downloader, the way the Python downloader lays it out",
		Args:  cobra.ExactArgs(1),
		RunE: func(command *cobra.Command, args []string) error {
			options.CourseURL = args[0]
			var err error
			options.DataDir, options.ChromeDataDir, err = withDataDefaults(options.DataDir, options.ChromeDataDir)
			if err != nil {
				return err
			}
			if options.OutputDir, err = filepath.Abs(options.OutputDir); err != nil {
				return fmt.Errorf("resolve --%s: %w", outputFlag, err)
			}
			options.ToolOutput = command.ErrOrStderr()
			logger := slog.New(slog.NewTextHandler(command.ErrOrStderr(), nil))
			summary, err := app.DownloadCourse(command.Context(), options, logger, func(event jobs.Event) {
				logger.Info("download progress", "event", event.Kind, "lecture", event.LectureTitle, "failures", event.Failures)
			})
			if err != nil {
				return err
			}
			output := command.OutOrStdout()
			fmt.Fprintf(output, "downloaded %d, already present %d, failed %d\n", summary.Downloaded, summary.AlreadyPresent, len(summary.Failed))
			for _, failure := range summary.Failed {
				fmt.Fprintf(output, "  failed: %s\n", failure)
			}
			return nil
		},
	}
	flags := command.Flags()
	flags.StringVar(&options.OutputDir, outputFlag, "", "directory the course folder is written to")
	flags.StringVar(&options.Chapters, "chapters", "", `chapters to download, like "1,3-5" (default: all)`)
	flags.StringVarP(&options.CaptionLanguage, "caption-language", "l", defaultCaptionLang, `caption language, "all", or "" for none`)
	flags.StringVar(&options.DataDir, "data-dir", "", dataDirUsage)
	flags.StringVar(&options.ChromeDataDir, "chrome-data-dir", "", chromeDataDirUsage)
	addToolFlags(command, &options.Tools)
	_ = command.MarkFlagRequired(outputFlag)
	return command
}
