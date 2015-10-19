// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package browse

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	livevim   bool
	blockname string
	startrune int
	stoprune  int
}

var browseCmd = &cobra.Command{
	Use:   "browse",
	Short: "lists all runes (constrained by block or range)",
	Run: func(cmd *cobra.Command, args []string) {
		srcs := sources.NewFast()
		if flags.livevim {
			srcs.LoadLiveVim()
		}

		if flags.stoprune == 0 {
			flags.stoprune = int(srcs.Unicode.MaxRune)
		}

		if flags.blockname != "" {
			begin, end, candidates := srcs.UBlocks.FindByName(flags.blockname)
			if end == 0 {
				fmt.Fprintf(os.Stderr, "unknown blockname %q\n", flags.blockname)
				if candidates != nil {
					fmt.Fprintf(os.Stderr, "there are %d possibilities:\n", len(candidates))
					for _, c := range candidates {
						fmt.Fprintf(os.Stderr, "\t%q\n", c)
					}
				}
				root.Errored()
				return
			}
			flags.startrune, flags.stoprune = int(begin), int(end)
		}

		results := resultset.New(srcs, int(flags.stoprune-flags.startrune)+100)
		lastBlock := ""

		stopAfter := rune(flags.stoprune)
		for r := rune(flags.startrune); r <= stopAfter; r++ {
			if ci, ok := srcs.Unicode.ByRune[r]; ok {
				if ci.Name == "<control>" {
					continue
				}
				block := srcs.UBlocks.Lookup(r)
				if block != lastBlock {
					if lastBlock != "" {
						results.AddDivider()
					}
					lastBlock = block
				}
				results.AddCharInfo(ci)
			}
		}

		//fmt.Printf("\ngot all input, asking to print tables now (%d/%d entries) ...\n", results.LenItemCount(), results.LenTotalCount())
		results.PrintTables()
	},
}

func init() {
	if !resultset.CanTable() {
		return
	}
	browseCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data (for verbose)")
	browseCmd.Flags().StringVarP(&flags.blockname, "block", "b", "", "only show this block")
	browseCmd.Flags().IntVarP(&flags.startrune, "from", "f", 0, "show range starting at this value")
	browseCmd.Flags().IntVarP(&flags.stoprune, "to", "t", 0, "show range ending at this value")
	root.AddCommand(browseCmd)
}
