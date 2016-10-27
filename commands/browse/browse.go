// Copyright © 2015,2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package browse

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/table"
	"github.com/philpennock/character/unicode"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	blockname  string
	limitAbort int
	listblocks bool
	livevim    bool
	startrune  int
	stoprune   int
	verbose    bool
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

		if flags.listblocks {
			showBlocks(srcs)
			return
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

		// We don't do limitAbort here, because there are gaps in the codepoint
		// assigments, and these lookups are fast.  The slowness (reason for
		// the limit) comes in termtables rendering.

		results := resultset.New(srcs, int(flags.stoprune-flags.startrune)+100)
		lastBlock := ""
		var firstCI, lastCI unicode.CharInfo

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
				lastCI = ci
				if firstCI.Number == 0 {
					firstCI = ci
				}
			}
		}

		//fmt.Printf("\ngot all input, asking to print tables now (%d/%d entries) ...\n", results.LenItemCount(), results.LenTotalCount())
		if results.LenItemCount() > flags.limitAbort {
			tmp := resultset.New(srcs, 3)
			tmp.OutputStream = os.Stderr
			tmp.AddCharInfo(firstCI)
			tmp.AddDivider()
			tmp.AddCharInfo(lastCI)

			fmt.Fprintf(os.Stderr,
				("that would show %d characters, more than %d limit\n" +
					"declining to proceed without -A override (gets slow)\n" +
					"Range covers start-end:\n"),
				results.LenItemCount(), flags.limitAbort)
			tmp.PrintTables()
			root.Errored()
			return
		}

		results.PrintTables()
	},
}

func showBlocks(srcs *sources.Sources) {
	if !flags.verbose {
		for _, blockInfo := range srcs.UBlocks.ListBlocks() {
			fmt.Printf("%s\n", blockInfo.Name)
		}
		return
	}
	// At time of writing, this is the only place outside of resultset
	// which is directly generating tables.  Probably still the right
	// thing to do.
	if !table.Supported() {
		fmt.Fprintf(os.Stderr, "sorry, this build is missing table support??\n")
		root.Errored()
		return
	}
	t := table.New()
	t.AddHeaders("Name", "From", "To")
	for _, blockInfo := range srcs.UBlocks.ListBlocks() {
		t.AddRow(
			blockInfo.Name,
			strconv.FormatUint(uint64(blockInfo.Min), 16),
			strconv.FormatUint(uint64(blockInfo.Max), 16),
		)
	}
	fmt.Print(t.Render())
}

func init() {
	if !resultset.CanTable() {
		return
	}
	browseCmd.Flags().StringVarP(&flags.blockname, "block", "b", "", "only show this block")
	browseCmd.Flags().IntVarP(&flags.limitAbort, "limit-abort", "A", 3000, "abort if would show more than this many entries")
	browseCmd.Flags().BoolVarP(&flags.listblocks, "list-blocks", "B", false, "list all available block names")
	browseCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data")
	browseCmd.Flags().IntVarP(&flags.startrune, "from", "f", 0, "show range starting at this value")
	browseCmd.Flags().IntVarP(&flags.stoprune, "to", "t", 0, "show range ending at this value")
	browseCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about block-names (for list-blocks)")
	root.AddCommand(browseCmd)
}
