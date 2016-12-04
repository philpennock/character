// Copyright © 2015,2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package root

import (
	"github.com/philpennock/character/table"

	"github.com/spf13/cobra"
)

var tableListStyles = &cobra.Command{
	Use:   "table-list-styles",
	Short: "list known table style options",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Printf("%s: table provider supports %d styles:\n", cmd.Name(), len(table.AvailableStyles))
		for _, style := range table.AvailableStyles {
			cmd.Printf("\t%q\n", style)
		}
	},
}

func init() {
	// flagSet: match the logic from root.go
	flagSet := characterCmd.PersistentFlags()
	if table.AvailableStyles != nil {
		flagSet.StringVar(&table.RenderStyle, "table-style", table.RenderStyle, "style option for table")
		AddCommand(tableListStyles)
	}
}
