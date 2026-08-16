package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var Version = "0.1.0"

func GetVersion() string {
	return "devdoc version " + Version + " " + runtime.GOOS + "/" + runtime.GOARCH
}

func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Prints the devdoc CLI version",
		Long:  "Prints the devdoc CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(os.Stdout, GetVersion())
			return nil
		},
	}
	return cmd
}
