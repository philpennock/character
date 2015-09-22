package version

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

// should be updated by the linker from the Go command-line, if make used
// for build
var VersionString string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version of character",
	Run: func(cmd *cobra.Command, args []string) {
		if VersionString == "" {
			VersionString = "<unknown>"
		}
		fmt.Printf("%s: version %s\n", cmd.Root().Name(), VersionString)
	},
}

func init() {
	root.AddCommand(versionCmd)
}
