// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode_test

import (
	"testing"

	"github.com/philpennock/character/unicode"
)

func TestGeneralCategory(t *testing.T) {
	tests := []struct {
		r    rune
		want string
		desc string
	}{
		{0x0041, "Lu", "LATIN CAPITAL LETTER A"},
		{0x0061, "Ll", "LATIN SMALL LETTER A"},
		{0x01C5, "Lt", "LATIN CAPITAL LETTER D WITH SMALL LETTER Z WITH CARON"},
		{0x02B0, "Lm", "MODIFIER LETTER SMALL H"},
		{0x4E00, "Lo", "CJK UNIFIED IDEOGRAPH-4E00"},
		{0x0300, "Mn", "COMBINING GRAVE ACCENT"},
		{0x0308, "Mn", "COMBINING DIAERESIS"},
		{0x0903, "Mc", "DEVANAGARI SIGN VISARGA"},
		{0x0030, "Nd", "DIGIT ZERO"},
		{0x2160, "Nl", "ROMAN NUMERAL ONE"},
		{0x00B2, "No", "SUPERSCRIPT TWO"},
		{0x005F, "Pc", "LOW LINE"},
		{0x002D, "Pd", "HYPHEN-MINUS"},
		{0x0028, "Ps", "LEFT PARENTHESIS"},
		{0x0029, "Pe", "RIGHT PARENTHESIS"},
		{0x00AB, "Pi", "LEFT-POINTING DOUBLE ANGLE QUOTATION MARK"},
		{0x00BB, "Pf", "RIGHT-POINTING DOUBLE ANGLE QUOTATION MARK"},
		{0x002E, "Po", "FULL STOP"},
		{0x002B, "Sm", "PLUS SIGN"},
		{0x0024, "Sc", "DOLLAR SIGN"},
		{0x005E, "Sk", "CIRCUMFLEX ACCENT"},
		{0x2713, "So", "CHECK MARK"},
		{0x0020, "Zs", "SPACE"},
		{0x0000, "Cc", "NULL"},
		{0x000A, "Cc", "LINE FEED"},
		{0x00AD, "Cf", "SOFT HYPHEN"},
		// Cn: a codepoint not assigned to any category
		{0xFDD0, "Cn", "noncharacter in Unicode — not in any table"},
	}
	for _, tt := range tests {
		got := unicode.GeneralCategory(tt.r)
		if got != tt.want {
			t.Errorf("GeneralCategory(U+%04X %s) = %q; want %q", tt.r, tt.desc, got, tt.want)
		}
	}
}
