package cmd

import (
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "devdoc",
		Short:         "A lightweight local markdown documentation preview tool",
		Long:          "A lightweight local markdown documentation preview tool",
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       Version,
	}

	cmd.AddCommand(NewBuildCommand())
	cmd.AddCommand(NewCleanCommand())
	cmd.AddCommand(NewInitCommand())
	cmd.AddCommand(NewRunCommand())
	cmd.AddCommand(NewVersionCommand())
	cmd.AddCommand(NewWatchCommand())

	return cmd
}
