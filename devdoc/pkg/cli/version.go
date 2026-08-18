package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var Version = "0.1.0"

// versionString returns the bare version + platform string used as
// cmd.Version. Cobra's `-v` / `--version` flag prepends the program name
// and " version " automatically, so this must NOT include "devdoc version"
// itself — otherwise the auto-prepended text is duplicated. The `version`
// subcommand below adds the prefix itself.
func versionString() string {
	return Version + " " + runtime.GOOS + "/" + runtime.GOARCH
}

func newVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the devdoc CLI version",
		Long:  "Prints the devdoc CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stdout, "devdoc version "+versionString())
			return nil
		},
	}
	return cmd
}
