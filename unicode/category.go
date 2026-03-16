// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"unicode"
)

// categoryEntry pairs a two-letter abbreviation with its Unicode range table.
type categoryEntry struct {
	abbr  string
	table *unicode.RangeTable
}

// orderedCategories is an ordered slice of Unicode general categories checked
// by GeneralCategory.  Order is significant: more-specific categories are
// tested before broader ones, and the order matches the Unicode standard.
var orderedCategories = []categoryEntry{
	// Letter
	{"Lu", unicode.Lu},
	{"Ll", unicode.Ll},
	{"Lt", unicode.Lt},
	{"Lm", unicode.Lm},
	{"Lo", unicode.Lo},
	// Mark
	{"Mn", unicode.Mn},
	{"Mc", unicode.Mc},
	{"Me", unicode.Me},
	// Number
	{"Nd", unicode.Nd},
	{"Nl", unicode.Nl},
	{"No", unicode.No},
	// Punctuation
	{"Pc", unicode.Pc},
	{"Pd", unicode.Pd},
	{"Ps", unicode.Ps},
	{"Pe", unicode.Pe},
	{"Pi", unicode.Pi},
	{"Pf", unicode.Pf},
	{"Po", unicode.Po},
	// Symbol
	{"Sm", unicode.Sm},
	{"Sc", unicode.Sc},
	{"Sk", unicode.Sk},
	{"So", unicode.So},
	// Separator
	{"Zs", unicode.Zs},
	{"Zl", unicode.Zl},
	{"Zp", unicode.Zp},
	// Other
	{"Cc", unicode.Cc},
	{"Cf", unicode.Cf},
	{"Cs", unicode.Cs},
	{"Co", unicode.Co},
}

// GeneralCategory returns the two-letter Unicode general category abbreviation
// for r (e.g. "Lu" for uppercase letters, "So" for other symbols).
// Returns "Cn" (unassigned) when no category table matches.
func GeneralCategory(r rune) string {
	for _, c := range orderedCategories {
		if unicode.Is(c.table, r) {
			return c.abbr
		}
	}
	return "Cn"
}
