// Copyright © 2016-2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package code

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/internal/encodings"
	"github.com/philpennock/character/internal/table"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	encodings   bool
	tableStyles bool
}

var knownCmd = &cobra.Command{
	Use:   "known",
	Short: "describes some internal lists of data",
	Run: func(cmd *cobra.Command, args []string) {
		doneSomething := false

		// We print results to stdout, not stderr, and cmd.Printf defaults to stderr

		if flags.encodings {
			fmt.Printf("%s %s: these encoding names (and some aliases) are known:\n",
				root.Cobra().Name(), cmd.Name())
			for _, enc := range encodings.ListKnownCharsets() {
				fmt.Printf("\t%q\n", enc)
			}
			doneSomething = true
		}

		if flags.tableStyles {
			fmt.Printf("%s %s: table provider supports %d styles:\n",
				root.Cobra().Name(), cmd.Name(), len(table.AvailableStyles))
			for _, style := range table.AvailableStyles {
				fmt.Printf("\t%q\n", style)
			}
			doneSomething = true
		}
		if !doneSomething {
			_ = cmd.Usage()
			root.Errored()
		}
	},
}

func init() {
	knownCmd.Flags().BoolVarP(&flags.encodings, "encodings", "e", false, "list known encodings/character-sets")
	knownCmd.Flags().BoolVarP(&flags.tableStyles, "table-styles", "t", false, "list known table styles")
	root.AddCommand(knownCmd)
}
