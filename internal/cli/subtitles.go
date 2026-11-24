package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/vukiman1/udemy-studio/internal/app"
)

const workFlag = "work"

func newSubtitlesCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "subtitles",
		Short: "Hand subtitles to a translator outside the app, such as a Claude Code session, and take the translations back",
	}
	command.AddCommand(newSubtitlesExportCommand(), newSubtitlesImportCommand())
	return command
}

func newSubtitlesExportCommand() *cobra.Command {
	var handoff app.SubtitleHandoff
	var course string
	command := &cobra.Command{
		Use:   "export",
		Short: "Write the English cues of every lecture without Vietnamese subtitles into the work dir, one <id>.en.txt each",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := absolutePaths(&handoff); err != nil {
				return err
			}
			count, err := handoff.Export(course)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(command.OutOrStdout(), "exported %d lectures to %s; write each translation as <id>.vi.txt beside <id>.en.txt\n", count, handoff.WorkDir)
			return err
		},
	}
	addHandoffFlags(command, &handoff)
	command.Flags().StringVar(&course, "course", "", "course folder name or id (default: every course)")
	return command
}

func newSubtitlesImportCommand() *cobra.Command {
	var handoff app.SubtitleHandoff
	command := &cobra.Command{
		Use:   "import",
		Short: "Check every <id>.vi.txt in the work dir as the app checks its own translations, and write those that pass into the library",
		Args:  cobra.NoArgs,
		RunE: func(command *cobra.Command, _ []string) error {
			if err := absolutePaths(&handoff); err != nil {
				return err
			}
			report, err := handoff.Import()
			if err != nil {
				return err
			}
			output := command.OutOrStdout()
			for _, section := range []struct {
				label string
				items []string
			}{{"written", report.Written}, {"rejected", report.Rejected}, {"not translated yet", report.Waiting}} {
				fmt.Fprintf(output, "%s: %d\n", section.label, len(section.items))
				for _, item := range section.items {
					fmt.Fprintf(output, "  %s\n", item)
				}
			}
			if len(report.Rejected) > 0 {
				return fmt.Errorf("%d translations were rejected", len(report.Rejected))
			}
			return nil
		},
	}
	addHandoffFlags(command, &handoff)
	return command
}

func addHandoffFlags(command *cobra.Command, handoff *app.SubtitleHandoff) {
	flags := command.Flags()
	flags.StringVar(&handoff.LibraryDir, "library", os.Getenv(libraryDirEnv), "downloads/ directory (env "+libraryDirEnv+")")
	flags.StringVar(&handoff.WorkDir, workFlag, "", "work directory holding the exported and translated text files")
	_ = command.MarkFlagRequired(workFlag)
}

func absolutePaths(handoff *app.SubtitleHandoff) error {
	if handoff.LibraryDir == "" {
		return app.ErrMissingLibraryDir
	}
	for _, path := range []*string{&handoff.LibraryDir, &handoff.WorkDir} {
		absolute, err := filepath.Abs(*path)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", *path, err)
		}
		*path = absolute
	}
	return nil
}
