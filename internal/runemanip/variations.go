// Copyright © 2022 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package runemanip

func IsVariationSelector(r rune) bool {
	// Block: Variation Selectors
	if r >= 0xFE00 && r <= 0xFE0F {
		return true
	}
	// Block: Variation Selectors Supplement
	if r >= 0xE0100 && r <= 0xE01EF {
		return true
	}
	return false
}
