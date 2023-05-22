// Copyright © 2023 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"testing"

	"github.com/liquidgecka/testlib"
)

func TestHaveGeneratedData(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()

	T.NotEqual(len(allKnownBlocks), 0, "have no known blocks")
	T.NotEqual(maxKnownBlockRune, 0, "have non-zero max known block rune")
	T.NotEqual(len(global.linearNames), 0, "have no linear list of codepoints")
	T.NotEqual(runeTotalCount, 0, "have runes")
	T.NotEqual(len(emojiable), 0, "have no emoji data")
	T.NotEqual(emojiableTotalCount, 0, "have codepoints which can be emojified")
}
