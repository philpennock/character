// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt
//
// Tables come from a manual translation of <https://xkcd.com/3054/>

package transform

import (
	"strings"

	"golang.org/x/text/unicode/norm"

	"github.com/spf13/cobra"
)

var UnicodeToScream map[rune]string
var AdditionalDecodes map[string]rune

func init() {
	// For the reverse direction, how do we deal with normalization forms and making sure we can handle any?
	UnicodeToScream = map[rune]string{
		'a': "a",  // LATIN SMALL LETTER A
		'b': "ȧ",  // LATIN SMALL LETTER A WITH DOT ABOVE
		'c': "a̧", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING CEDILLA' (I read it as cedilla, not ogonek)
		'd': "a̱", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING MACRON BELOW'
		'e': "á",  // LATIN SMALL LETTER A WITH ACUTE (or could be: character named -1c 'LATIN SMALL LETTER A' 'COMBINING ACUTE ACCENT')
		'f': "a̮", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING BREVE BELOW'
		'g': "a̋", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING DOUBLE ACUTE ACCENT'
		'h': "a̰", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING TILDE BELOW'
		'i': "ả",  // LATIN SMALL LETTER A WITH HOOK ABOVE  (the COMBINING HOOK ABOVE form did not work in my terminal)
		'j': "a̓", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING COMMA ABOVE'
		'k': "ạ",  // LATIN SMALL LETTER A WITH DOT BELOW
		'l': "ă",  // LATIN SMALL LETTER A WITH BREVE
		'm': "ǎ",  // LATIN SMALL LETTER A WITH CARON
		'n': "â",  // LATIN SMALL LETTER A WITH CIRCUMFLEX
		'o': "å",  // LATIN SMALL LETTER A WITH RING ABOVE
		'p': "a̭", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING CIRCUMFLEX ACCENT BELOW'
		'q': "a̤", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING DIAERESIS BELOW'
		'r': "ȃ", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING INVERTED BREVE'
		's': "ã",  // LATIN SMALL LETTER A WITH TILDE
		't': "ā",  // LATIN SMALL LETTER A WITH MACRON
		'u': "ä",  // LATIN SMALL LETTER A WITH DIAERESIS
		'v': "à",  // LATIN SMALL LETTER A WITH GRAVE
		'w': "ȁ",  // LATIN SMALL LETTER A WITH DOUBLE GRAVE
		'x': "aͯ", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING LATIN SMALL LETTER X'
		'y': "a̦", // character named -1c 'LATIN SMALL LETTER A' 'COMBINING COMMA BELOW' (renders the comma far to the right for me
		'z': "ⱥ",  // LATIN SMALL LETTER A WITH STROKE

		'A': "A",  // LATIN CAPITAL LETTER A
		'B': "Ȧ",  // LATIN CAPITAL LETTER A WITH DOT ABOVE
		'C': "A̧", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING CEDILLA' (I read it as cedilla, not ogonek)
		'D': "A̱", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING MACRON BELOW'
		'E': "Á",  // LATIN CAPITAL LETTER A WITH ACUTE (or could be: character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING ACUTE ACCENT')
		'F': "A̮", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING BREVE BELOW'
		'G': "A̋", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING DOUBLE ACUTE ACCENT'
		'H': "A̰", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING TILDE BELOW'
		'I': "Ả",  // LATIN CAPITAL LETTER A WITH HOOK ABOVE  (the COMBINING HOOK ABOVE form did not work in my terminal)
		'J': "A̓", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING COMMA ABOVE'
		'K': "Ạ",  // LATIN CAPITAL LETTER A WITH DOT BELOW
		'L': "Ă",  // LATIN CAPITAL LETTER A WITH BREVE
		'M': "Ǎ",  // LATIN CAPITAL LETTER A WITH CARON
		'N': "Â",  // LATIN CAPITAL LETTER A WITH CIRCUMFLEX
		'O': "Å",  // LATIN CAPITAL LETTER A WITH RING ABOVE
		'P': "A̭", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING CIRCUMFLEX ACCENT BELOW'
		'Q': "A̤", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING DIAERESIS BELOW'
		'R': "Ȃ", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING INVERTED BREVE'
		'S': "Ã",  // LATIN CAPITAL LETTER A WITH TILDE
		'T': "Ā",  // LATIN CAPITAL LETTER A WITH MACRON
		'U': "Ä",  // LATIN CAPITAL LETTER A WITH DIAERESIS
		'V': "À",  // LATIN CAPITAL LETTER A WITH GRAVE
		'W': "Ȁ",  // LATIN CAPITAL LETTER A WITH DOUBLE GRAVE
		'X': "Aͯ", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING LATIN SMALL LETTER X'
		'Y': "A̦", // character named -1c 'LATIN CAPITAL LETTER A' 'COMBINING COMMA BELOW' (renders the comma far to the right for me
		'Z': "Ⱥ",  // LATIN CAPITAL LETTER A WITH STROKE
	}

	AdditionalDecodes = map[string]rune{
		// Wikifunctions has also introduced an encoder, at <https://www.wikifunctions.org/view/en/Z22725>
		// Feeding it `The quick brown fox jumped over the lazy dog` it converted to upper-case and when decoding,
		// we hit one missed decode:
		"A̯": 'P', // LATIN CAPITAL LETTER A, COMBINING INVERTED BREVE BELOW
		// Looking at <https://xkcd.com/3054/> 'P' closely, there is a sharp point so I'm sticking to my interpretation as a circumflex, not an inverted breve.
	}
}

var screamOptions struct {
	wantDecode bool
}

type StreamCoder struct {
	replacer *strings.Replacer
}

func (sc *StreamCoder) Replace(s string) string { return sc.replacer.Replace(s) }

func NewEncoder() *StreamCoder {
	replacerArgs := make([]string, 0, 2*len(UnicodeToScream))
	for Rune, Scream := range UnicodeToScream {
		replacerArgs = append(replacerArgs, string(Rune), norm.NFC.String(Scream))
	}
	return &StreamCoder{
		replacer: strings.NewReplacer(replacerArgs...),
	}
}

func NewDecoder() *StreamCoder {
	replacerArgs := make([]string, 0, 8*len(UnicodeToScream))
	for Rune, Scream := range UnicodeToScream {
		bare := string(Rune)
		// without the t!=bare check, the decode will see A->A (0x41 -> 0x41) and this will preempt other longer strings
		for _, f := range []func(string) string{norm.NFC.String, norm.NFD.String, norm.NFKC.String, norm.NFKD.String} {
			t := f(Scream)
			if t != bare {
				replacerArgs = append(replacerArgs, t, bare)
			}
		}
	}
	for Scream, Rune := range AdditionalDecodes {
		bare := string(Rune)
		for _, f := range []func(string) string{norm.NFC.String, norm.NFD.String, norm.NFKC.String, norm.NFKD.String} {
			t := f(Scream)
			if t != bare {
				replacerArgs = append(replacerArgs, t, bare)
			}
		}
	}
	return &StreamCoder{
		replacer: strings.NewReplacer(replacerArgs...),
	}
}

var screamSubCommand = transformCobraCommand{
	Use:   "scream",
	Short: "toggle characters between plain & Scream (XKCD 3054)",
	DoFlags: func(c *cobra.Command) {
		c.PersistentFlags().BoolVarP(&screamOptions.wantDecode, "decode", "d", false, "decode screams, not encode")
	},
	Transformer: func(args []string) (string, error) {
		var coder *StreamCoder
		if len(args) == 0 {
			return "", nil
		}
		if screamOptions.wantDecode {
			coder = NewDecoder()
		} else {
			coder = NewEncoder()
		}
		output := make([]string, len(args))
		for argI := range args {
			output[argI] = coder.Replace(args[argI])
		}
		return strings.Join(output, " "), nil
	},
}
