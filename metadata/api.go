// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package metadata

import (
	"fmt"

	"github.com/philpennock/character/unicode"
)

func IsPairCode(r rune) bool {
	// Enclosed Alphanumeric Supplement Block
	// "REGIONAL INDICATOR SYMBOL LETTER A" -- "REGIONAL INDICATOR SYMBOL LETTER Z"
	if r >= 0x1F1E6 && r <= 0x1F1FF {
		return true
	}
	return false
}

func PairCharInfo(r1, r2 rune) (unicode.CharInfo, bool) {
	return unicode.CharInfo{
		Number: 0,
		Name:   fmt.Sprintf("%c%c  - %s + %s", r1, r2, labelOf(r1), labelOf(r2)),
	}, true
}

func labelOf(r rune) string {
	if r >= 0x1F1E6 && r <= 0x1F1FF {
		return fmt.Sprintf("Regional Indicator %c", r-0x1F1E6+'A')
	}
	return "<unknown>"
}
