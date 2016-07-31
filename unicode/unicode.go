// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"bytes"
	"io"
	"sync"

	"github.com/philpennock/character/aux"

	"github.com/argusdusty/Ferret"
)

// CharInfo is the basic set of information about one Unicode character.
// We record the codepoint (as a Go rune) and the formal Name.
// We also store property data to let us know the display width.
type CharInfo struct {
	_      struct{}
	Number rune
	Name   string
	width  int
}

// TerminalCellWidth is how wide a character is when displayed in a cell grid.
// Unfortunately, full-width characters do not have this encoded as a property
// in a column; instead the Character Decomposition Mapping starts `<wide>`.
// This is a "tag".
//
// So we need to look at multiple columns from the source data, so might as well
// collect it in one parse.
// FIXME: This should move into generation-time analysis instead of storing the
// raw unicode-data.
func (ci CharInfo) TerminalCellWidth() int {
	return ci.width
}

// Unicode is the set of all data about all characters which we've retrieved
// from formal Unicode specifications.
type Unicode struct {
	ByRune  map[rune]CharInfo
	ByName  map[string]CharInfo
	Search  *ferret.InvertedSuffix
	MaxRune rune

	// these will be blanked once setup is complete
	linearNames   []string
	linearIfaceCI []interface{}
}

var global Unicode
var once struct {
	parseUnicode sync.Once
	loadSearch   sync.Once
}

// Load gives us all the Unicode-spec derived data which we have.
func Load() Unicode {
	once.parseUnicode.Do(parseRaw)
	return global
}

// LoadSearch gives us all the Unicode data, with search too; the search
// loading is slow, so we skip it by default.
func LoadSearch() Unicode {
	once.parseUnicode.Do(parseRaw)
	once.loadSearch.Do(addSearch)
	return global
}

func parseRaw() {
	b := bytes.NewBuffer(rawData)

	byRune := make(map[rune]CharInfo, rawLineCount)
	byName := make(map[string]CharInfo, rawLineCount)
	linearNames := make([]string, 0, rawLineCount)
	linearIfaceCI := make([]interface{}, 0, rawLineCount)
	var max rune

	lineNum := 0
	for {
		if b.Len() == 0 {
			break
		}
		line, err := b.ReadBytes('\n')
		lineNum++
		if err != nil {
			switch err {
			case io.EOF:
				break
			default:
				panic(err.Error())
			}
		}
		line = line[:len(line)-1]

		// our embedding inserts an extra newline at the start; be resistant
		if len(line) == 0 {
			continue
		}

		fields := bytes.FieldsFunc(line, func(r rune) bool { return r == ';' })

		r := aux.RuneFromHexField(fields[0])
		name := string(fields[1])
		ci := CharInfo{
			Number: r,
			Name:   name,
			width:  1,
		}
		if len(fields[2]) == 2 && fields[2][0] == 'M' {
			if fields[2][1] == 'n' || fields[2][1] == 'c' {
				// Mn or Mc for General Category
				ci.width = 0
			}
		}
		if ci.width == 1 {
			if bytes.HasPrefix(fields[5], []byte("<wide>")) {
				ci.width = 2
			} else if bytes.HasPrefix(fields[5], []byte("<vertical>")) {
				// XXX: is this normal or just how it is for my terminal/font?
				ci.width = 2
			}
		}
		byRune[r] = ci
		byName[name] = ci
		linearNames = append(linearNames, name)
		linearIfaceCI = append(linearIfaceCI, ci)
		if r > max {
			max = r
		}
	}

	global = Unicode{
		ByRune:        byRune,
		ByName:        byName,
		MaxRune:       max,
		linearNames:   linearNames,
		linearIfaceCI: linearIfaceCI,
	}
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
