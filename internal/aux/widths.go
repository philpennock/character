// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package aux

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// We need block info from unicode, which depends upon us.  Bleh.
var OverrideWidthMSPMin, OverrideWidthMSPMax rune // Miscellaneous Symbols and Pictographs

// DisplayCellWidth is a best-guess at how many "terminal grid cells" wide a character is
// The actual calculations are done by the table layer; our preferred table layer
// (tabular; because we wrote it and it's better) uses github.com/mattn/go-runewidth and this
// same function.
func DisplayCellWidth(s string) (width int, isOverride bool) {
	l := utf8.RuneCountInString(s)
	switch l {
	case 0:
		return 0, false
	case 1:
		break
	default:
		return runewidth.StringWidth(s), false
	}
	r, _ := utf8.DecodeRuneInString(s)
	switch {
	case r == utf8.RuneError:
		return 1, true
	case IsPairCode(r):
		return 1, true
	case OverrideWidthMSPMin != 0 && OverrideWidthMSPMin <= r && r <= OverrideWidthMSPMax:
		return 1, true
	default:
		return runewidth.StringWidth(s), false
	}
}
