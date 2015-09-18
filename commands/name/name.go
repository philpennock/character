package name

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/unicode"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	verbose bool
}

var nameCmd = &cobra.Command{
	Use:   "name [char... [char...]]",
	Short: "shows information about supplied characters",
	Run: func(cmd *cobra.Command, args []string) {
		u := unicode.Load()
		approxCharCount := 0
		for _, a := range args {
			approxCharCount += len(a) + 1
		}
		results := resultset.New(approxCharCount)

		for i, arg := range args {
			if i > 0 {
				results.AddDivider()
			}
			for _, r := range arg {
				if ci, ok := u.ByRune[r]; ok {
					results.AddCharInfo(ci)
				} else {
					root.Errored()
					// FIXME: proper error type
					results.AddError(string(r), fmt.Errorf("unknown codepoint %x", int(r)))
				}
			}
		}

		if flags.verbose {
			results.PrintTables()
		} else {
			results.PrintPlain(resultset.PRINT_NAME)
		}
	},
}

func init() {
	nameCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about the character")
	root.AddCommand(nameCmd)
}
