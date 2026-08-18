package cmd

import (
	"github.com/qhai-dev/devdoc/internal/runner"

	"github.com/spf13/cobra"
)

// NewBuildCommand implements the `devdoc build` subcommand: render the
// book into its build directory. It mirrors src/cmd/build.rs.
func NewBuildCommand() *cobra.Command {
	var dir, dest string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "build a book",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBuild(dir, dest)
		},
	}

	cmd.Flags().StringVar(&dir, "dir", ".", "book root")
	cmd.Flags().StringVar(&dest, "dest-dir", "", "output directory (overrides devdoc.yaml)")

	return cmd
}

func runBuild(dir, dest string) error {
	m, err := runner.Load(dir)
	if err != nil {
		return err
	}
	if dest != "" {
		m.Config.Build.BuildDir = dest
	}
	return m.Build()
}
