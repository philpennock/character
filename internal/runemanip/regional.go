// Copyright © 2016-2017,2021 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package runemanip

// IsPairCode indicates if the passed run is part of a pairing code designed
// for extensible lookup.  This is used for national flags.
// Eg, [0x1F1FA 0x1F1F8] = [<Regional Indicator U> <Regional Indicator S>] = 🇺🇸
// which will display in some contexts as the flag of the USA.
func IsPairCode(r rune) bool {
	// Enclosed Alphanumeric Supplement Block
	// "REGIONAL INDICATOR SYMBOL LETTER A" -- "REGIONAL INDICATOR SYMBOL LETTER Z"
	if r >= 0x1F1E6 && r <= 0x1F1FF {
		return true
	}
	return false
}
