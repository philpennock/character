// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"sync"

	"github.com/argusdusty/Ferret"
)

// CharInfo is the basic set of information about one Unicode character.
// We record the codepoint (as a Go rune) and the formal Name.
type CharInfo struct {
	_         struct{}
	Number    rune
	Name      string
	NameWidth int // occasional override
}

// Unicode is the set of all data about all characters which we've retrieved
// from formal Unicode specifications.
type Unicode struct {
	ByRune  map[rune]CharInfo
	ByName  map[string]CharInfo
	Search  *ferret.InvertedSuffix
	MaxRune rune

	// these will be blanked by Search once Search setup is complete
	linearNames   []string
	linearIfaceCI []interface{}
}

var once struct {
	loadSearch sync.Once
}

// Load gives us all the Unicode-spec derived data which we have.
func Load() Unicode {
	return global
}

// LoadSearch gives us all the Unicode data, with search too; the search
// loading is slow, so we skip it by default.
func LoadSearch() Unicode {
	once.loadSearch.Do(addSearch)
	return global
}

func addSearch() {
	global.Search = ferret.New(
		global.linearNames,
		global.linearNames,
		global.linearIfaceCI,
		ferret.UnicodeToLowerASCII)

	global.linearNames = nil
	global.linearIfaceCI = nil
}
