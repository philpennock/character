// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package name

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/metadata"
	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	livevim bool
	verbose bool
}

var nameCmd = &cobra.Command{
	Use:   "name [char... [char...]]",
	Short: "shows information about supplied characters",
	Run: func(cmd *cobra.Command, args []string) {
		srcs := sources.NewFast()
		if flags.verbose && flags.livevim {
			srcs.LoadLiveVim()
		}
		approxCharCount := 0
		for _, a := range args {
			approxCharCount += len(a) + 1
		}
		results := resultset.New(srcs, approxCharCount)

		var pairedCodepoint rune = 0

		for i, arg := range args {
			if i > 0 {
				results.AddDivider()
			}
			pairedCodepoint = 0
			for _, r := range arg {
				if ci, ok := srcs.Unicode.ByRune[r]; ok {
					results.AddCharInfo(ci)
					// Ancillary extra data if warranted
					if metadata.IsPairCode(ci.Number) {
						if pairedCodepoint != 0 {
							if ci2, ok := metadata.PairCharInfo(pairedCodepoint, ci.Number); ok {
								results.AddCharInfo(ci2)
							} else {
								results.AddError("", fmt.Errorf("unknown codepair %x-%x", pairedCodepoint, ci.Number))
							}
							pairedCodepoint = 0
						} else {
							pairedCodepoint = ci.Number
						}
					}
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
	if resultset.CanTable() {
		nameCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data (for verbose)")
		nameCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about the character")
	}
	// FIXME: support verbose results without tables

	root.AddCommand(nameCmd)
}
