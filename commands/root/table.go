// Copyright © 2015-2017,2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package root

import (
	"github.com/philpennock/character/internal/table"
)

func init() {
	// flagSet: match the logic from root.go
	flagSet := characterCmd.PersistentFlags()
	if table.AvailableStyles != nil {
		flagSet.StringVar(&table.RenderStyle, "table-style", table.RenderStyle, "style option for table")
	}
}

func tableRequestedList() bool {
	switch table.RenderStyle {
	case "", "?", "help":
		return true
	default:
		return false
	}
}
