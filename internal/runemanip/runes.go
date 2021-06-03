// Copyright © 2015,2021 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package runemanip

// RuneFromHexField converts a hex field, perhaps of odd length, to a rune
func RuneFromHexField(bb []byte) rune {
	// fields[0] is the hex encoding, but with perhaps odd numbers of bytes present (eg, 5)
	// So rather than `hex.Decode()`, we just decode manually
	var r rune
	for _, c := range bb {
		r *= 16
		switch {
		case '0' <= c && c <= '9':
			r += rune(c - '0')
		case 'A' <= c && c <= 'F':
			r += rune(10 + c - 'A')
		case 'a' <= c && c <= 'f':
			r += rune(10 + c - 'a')
		}
	}
	return r
}
