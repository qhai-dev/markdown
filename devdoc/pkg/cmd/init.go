package cmd

import (
	"github.com/qhai-dev/devdoc/internal/runner"

	"github.com/spf13/cobra"
)

// NewInitCommand implements the `devdoc init` subcommand: create a
// new book skeleton. It mirrors src/cmd/init.rs.
func NewInitCommand() *cobra.Command {
	var dir string
	var title string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "create a new book skeleton",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runner.Init(dir, runner.InitOptions{Title: title})
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "book root")
	cmd.Flags().StringVar(&title, "title", "", "book title (default \"My Book\" if empty)")

	return cmd
}
