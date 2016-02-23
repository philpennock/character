// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package fraktur

import (
	"fmt"
	"testing"

	"github.com/liquidgecka/testlib"
)

func TestRuneMapping(t *testing.T) {
	T := testlib.NewT(t)

	for _, pair := range []struct{ from, to rune }{
		{'A', '𝔄'},
		{'C', 'ℭ'},
		{'H', 'ℌ'},
		{'I', 'ℑ'},
		{'J', '𝔍'},
		{'R', 'ℜ'},
		{'Y', '𝔜'},
		{'Z', 'ℨ'},
		{'a', '𝔞'},
		{'c', '𝔠'},
		{'d', '𝔡'},
		{'z', '𝔷'},
		{'𝔄', 'A'},
		{'𝔍', 'J'},
		{'𝔜', 'Y'},
		{'𝔞', 'a'},
		{'𝔠', 'c'},
		{'𝔡', 'd'},
		{'𝔷', 'z'},
		{'𝕬', 'A'},
		{'𝕮', 'C'},
		{'𝕳', 'H'},
		{'𝕴', 'I'},
		{'𝕵', 'J'},
		{'𝕽', 'R'},
		{'𝖄', 'Y'},
		{'𝖅', 'Z'},
		{'𝖆', 'a'},
		{'𝖈', 'c'},
		{'𝖉', 'd'},
		{'𝖟', 'z'},
		{'ℭ', 'C'},
		{'ℌ', 'H'},
		{'ℑ', 'I'},
		{'ℜ', 'R'},
		{'ℨ', 'Z'},
	} {
		T.Equal(toggleRune(pair.from), pair.to, fmt.Sprintf("fraktur rune mapping equality map(%c)->%c", pair.from, pair.to))
	}
}
