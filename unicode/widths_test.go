// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

// We need the unicode init() setting of override vars for widths for this to work.

import (
	"fmt"
	"testing"

	"github.com/liquidgecka/testlib"

	"github.com/philpennock/character/internal/aux"
)

func TestDisplayCellWidth(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()

	for _, tuple := range []struct {
		in        string
		needWidth int
	}{
		{"", 0},
		{"a", 1},
		{"🤞", 1}, // Supplemental Symbols and Pictographs
		{"🌮", 1}, // Miscellaneous Symbols and Pictographs
		{"€", 1},
		{"☺", 1},
		{"😇", 1},
	} {
		haveWidth, _ := aux.DisplayCellWidth(tuple.in)
		T.Equal(haveWidth, tuple.needWidth, fmt.Sprintf("width of %q should be %d but got %d", tuple.in, tuple.needWidth, haveWidth))
	}
}
