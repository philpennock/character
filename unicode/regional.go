// Copyright © 2016-2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"fmt"

	"github.com/philpennock/character/aux"
)

// PairCharInfo returns a faked-up CharInfo which is for rune 0 but with an informative name.
func PairCharInfo(r1, r2 rune) (CharInfo, bool) {
	name := fmt.Sprintf("%c%c  - %s + %s", r1, r2, labelOf(r1), labelOf(r2))
	// aux.DisplayCellWidth decays to regular rune-based counting for lengths greater than 1
	width, _ := aux.DisplayCellWidth(name)
	// In testing, the go-runewidth based stuff does not handle regional indicators.
	width -= 3

	return CharInfo{
		Number:    0,
		Name:      name,
		NameWidth: width,
	}, true
}

func labelOf(r rune) string {
	if r >= 0x1F1E6 && r <= 0x1F1FF {
		return fmt.Sprintf("Regional Indicator %c", r-0x1F1E6+'A')
	}
	return "<unknown>"
}
