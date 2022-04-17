// Copyright © 2017,2021 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package runemanip

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// We need block info from unicode, which depends upon us.  Bleh.
// NOTE: this means that testing this requires importing unicode too, so the tests
// are elsewhere.  Eww.
var (
	OverrideWidthMSPMin, OverrideWidthMSPMax             rune // Miscellaneous Symbols and Pictographs
	OverrideWidthSSPMin, OverrideWidthSSPMax             rune // Supplemental Symbols and Pictographs
	OverrideWidthEmoticonsMin, OverrideWidthEmoticonsMax rune // Emoticons
)

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
	case IsVariationSelector(r):
		return 0, true
	case OverrideWidthMSPMin != 0 && OverrideWidthMSPMin <= r && r <= OverrideWidthMSPMax:
		fallthrough
	case OverrideWidthSSPMin != 0 && OverrideWidthSSPMin <= r && r <= OverrideWidthSSPMax:
		fallthrough
	case OverrideWidthEmoticonsMin != 0 && OverrideWidthEmoticonsMin <= r && r <= OverrideWidthEmoticonsMax:
		return 2, true
	default:
		return runewidth.StringWidth(s), false
	}
}
