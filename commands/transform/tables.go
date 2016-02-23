// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package transform

import (
	"sync"
	"unicode/utf8"
)

var updateTablesLock sync.RWMutex

func makeBidirectionalRuneTableUnsafe(charsA, charsB string) map[rune]rune {
	if utf8.RuneCountInString(charsA) != utf8.RuneCountInString(charsB) {
		panic("hard error in source mismatched lengths of string replacements")
	}
	table := make(map[rune]rune, 2*len(charsA))
	var ri, rj rune
	for i, j, wi, wj := 0, 0, 0, 0; i < len(charsA); {
		ri, wi = utf8.DecodeRuneInString(charsA[i:])
		rj, wj = utf8.DecodeRuneInString(charsB[j:])
		if ri != ' ' && rj != ' ' {
			table[ri] = rj
			table[rj] = ri
		}
		i += wi
		j += wj
	}
	return table
}

// assumes not updated, only ever created once
func ensureBidirectionalRuneTable(table *map[rune]rune, charsA, charsB string) {
	updateTablesLock.RLock()
	if *table != nil {
		updateTablesLock.RUnlock()
		return
	}
	updateTablesLock.RUnlock()
	updateTablesLock.Lock()
	defer updateTablesLock.Unlock()
	if *table != nil {
		return
	}
	*table = makeBidirectionalRuneTableUnsafe(charsA, charsB)
}

type transformFlags int

const (
	TRANS_REVERSE transformFlags = 1 << iota
)

func transformText(table map[rune]rune, args []string, flags transformFlags) string {
	text := make([]rune, 0, 2000)
	for argCount, arg := range args {
		if argCount > 0 {
			if t, found := table[' ']; found {
				text = append(text, t)
			} else {
				text = append(text, ' ')
			}
		}
		for _, r := range arg {
			if t, found := table[r]; found {
				text = append(text, t)
			} else {
				text = append(text, r)
			}
		}
	}

	if (flags&TRANS_REVERSE) != 0 && len(text) >= 2 {
		last := len(text) - 1
		for i := 0; i <= (last)/2; i++ {
			text[i], text[last-i] = text[last-i], text[i]
		}
	}

	// don't put a newline in here, the caller won't want that in clipboard insertion

	return string(text)
}
