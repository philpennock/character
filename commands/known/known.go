// Copyright © 2016-2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package known

import (
	"strconv"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
	"github.com/philpennock/character/internal/encodings"
	"github.com/philpennock/character/internal/table"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"
)

var flags struct {
	blocks        bool
	blocksVerbose bool
	encodings     bool
	nameOnly      bool
	tableStyles   bool
	verbose       bool // not the resultset version, _only_ want -v
}

var knownCmd = &cobra.Command{
	Use:   "known",
	Short: "describes some internal lists of data",
	Run: func(cmd *cobra.Command, args []string) {
		doneSomething := false

		// Compatibility with old `character browse -B`
		if flags.blocksVerbose {
			flags.blocks = true
			if !flags.nameOnly {
				flags.verbose = true
			}
		}

		if flags.verbose && flags.nameOnly {
			root.Errorf("incompatible flags, not verbose and name-only\n")
			return
		}

		Fill := func(l *lister) *lister {
			l.command = cmd.Name()
			l.nameOnly = flags.nameOnly
			l.verbose = flags.verbose
			return l
		}

		if flags.encodings {
			Fill(&lister{
				nonTableLabel: "know %d encoding names (and some more aliases)",
				columnTitles:  []interface{}{"Encoding"},
			}).Each(encodings.ListKnownCharsets())
			doneSomething = true
		}

		if flags.tableStyles {
			Fill(&lister{
				nonTableLabel: "table provider supports %d styles",
				columnTitles:  []interface{}{"Style"},
			}).Each(table.AvailableStyles)
			doneSomething = true
		}

		if flags.blocks {
			Fill(&lister{
				nonTableLabel: "know of %d Unicode block names",
				columnTitles:  []interface{}{"Name", "From", "To"},
				fieldsExtract: func(row interface{}) []interface{} {
					bi := row.(unicode.BlockInfo)
					return []interface{}{
						bi.Name,
						strconv.FormatUint(uint64(bi.Min), 16),
						strconv.FormatUint(uint64(bi.Max), 16),
					}
				},
			}).Each(sources.NewFast().UBlocks.ListBlocks())
			doneSomething = true
		}

		if !doneSomething {
			_ = cmd.Usage()
			root.Errored()
		}
	},
}

func init() {
	knownCmd.Flags().BoolVarP(&flags.blocks, "blocks", "b", false, "list known Unicode block names")
	// -B for compat with the same functionality in `browse` cmd
	knownCmd.Flags().BoolVarP(&flags.blocksVerbose, "blocks__verbose", "B", false, "list known Unicode block names")
	knownCmd.Flags().MarkHidden("blocks__verbose")
	knownCmd.Flags().BoolVarP(&flags.encodings, "encodings", "e", false, "list known encodings/character-sets")
	knownCmd.Flags().BoolVarP(&flags.tableStyles, "table-styles", "t", false, "list known table styles")
	knownCmd.Flags().BoolVarP(&flags.nameOnly, "name-only", "n", false, "list only names, unquoted")
	knownCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information in tables, possibly multi-column")
	root.AddCommand(knownCmd)
}
