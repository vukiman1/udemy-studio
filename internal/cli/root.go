package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

const (
	exitOK      = 0
	exitFailure = 1
)

func Execute(version string) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := newRootCommand(version).ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "udemy-studio:", err)
		return exitFailure
	}
	return exitOK
}

func newRootCommand(version string) *cobra.Command {
	root := &cobra.Command{
		Use:           "udemy-studio",
		Short:         "Offline Udemy course library and player",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newServeCommand(version), newPlanCommand(), newVersionCommand(version))
	return root
}
