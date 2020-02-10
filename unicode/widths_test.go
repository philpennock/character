// Copyright © 2017,2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

// We need the unicode init() setting of override vars for widths for this to work.

import (
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/liquidgecka/testlib"

	"github.com/philpennock/character/internal/aux"
)

func TestDisplayCellWidth(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()

	for itemNum, tuple := range []struct {
		in        string
		needWidth int
	}{
		{"", 0},
		{"a", 1},
		{"🤞", 2}, // Supplemental Symbols and Pictographs
		{"🌮", 2}, // Miscellaneous Symbols and Pictographs
		{"€", 1},
		{"☺", 1},
		{"😇", 2},
	} {
		haveWidth, _ := aux.DisplayCellWidth(tuple.in)
		testedRune, _ := utf8.DecodeRuneInString(tuple.in)
		T.Equal(haveWidth, tuple.needWidth, fmt.Sprintf("test %d: width of %q [%#x] should be %d but got %d", itemNum, tuple.in, testedRune, tuple.needWidth, haveWidth))
	}
}
