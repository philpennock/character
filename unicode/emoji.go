// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"github.com/philpennock/character/internal/aux"
)

// Emojiable indicates whether or not a given rune might be an emoji and so can
// be followed by a presentation selector, per UTS#51 on Unicode Emoji.
// Various characters can be followed by 0xFE0E or 0xFE0F to select text or
// emoji variants and override normal rendering, but this is only well defined
// for the characters in a Consortium-maintained table.  In at least one
// terminal emulator, _some_ pictograms followed by the emoji presentation
// selector will emit garbage sequences instead of just the base pictogram.
//
// HOWEVER: various emojis from the Emoticons block can be given the text
// selector and it "works" in iTerm on macOS, but this list is not complete and
// so simple presence-in-generated-list is not sufficient.
// Going on gut (so probably wrong) am using the same list of "stuff in these
// blocks is probably not marked as the right width" as we use elsewhere.
func Emojiable(r rune) bool {
	_, ok := emojiable[r]
	if ok {
		return ok
	}
	if aux.OverrideWidthEmoticonsMin <= r && r <= aux.OverrideWidthEmoticonsMax {
		return true
	}
	if aux.OverrideWidthMSPMin <= r && r <= aux.OverrideWidthMSPMax {
		return true
	}
	if aux.OverrideWidthSSPMin <= r && r <= aux.OverrideWidthSSPMax {
		return true
	}
	return false
}
