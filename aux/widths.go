// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package aux

import (
	"github.com/mattn/go-runewidth"
)

// DisplayCellWidth is a best-guess at how many "terminal grid cells" wide a character is
// The actual calculations are done by the table layer; our preferred table layer
// (tabular; because we wrote it and it's better) uses github.com/mattn/go-runewidth and this
// same function.
func DisplayCellWidth(s string) int {
	return runewidth.StringWidth(s)
}
