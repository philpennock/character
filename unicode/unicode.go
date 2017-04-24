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
	linearNames []string
	linearCI    []CharInfo
}

var once struct {
	populateUnicode sync.Once
	loadSearch      sync.Once
}

// Load gives us all the Unicode-spec derived data which we have.
func Load() Unicode {
	once.populateUnicode.Do(populateUnicode)
	return global
}

// LoadSearch gives us all the Unicode data, with search too; the search
// loading is slow, so we skip it by default.
func LoadSearch() Unicode {
	once.loadSearch.Do(addSearch)
	return global
}

func addSearch() {
	linearIfaceCI := make([]interface{}, len(global.linearCI))
	for i := range global.linearCI {
		linearIfaceCI[i] = global.linearCI[i]
	}
	global.Search = ferret.New(
		global.linearNames,
		global.linearNames,
		linearIfaceCI,
		ferret.UnicodeToLowerASCII)

	global.linearNames = nil
	global.linearCI = nil
}

// populateUnicode is done so that we don't need to put the maps into the
// generated code, because that triggered some rather unfortunate degenerate
// performance (40s to build instead of 1s).
func populateUnicode() {
	global.ByRune = make(map[rune]CharInfo, runeTotalCount)
	global.ByName = make(map[string]CharInfo, runeTotalCount)

	for i := range global.linearCI {
		ci := &global.linearCI[i]
		global.ByRune[ci.Number] = *ci
		// skip the <control> dups
		if ci.Number >= 32 {
			global.ByName[ci.Name] = *ci
		}
	}
}
