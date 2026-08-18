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
		Version:       versionString(),
	}

	cmd.AddCommand(newBuildCommand())
	cmd.AddCommand(newCleanCommand())
	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newRunCommand())
	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newWatchCommand())

	return cmd
}
