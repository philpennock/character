// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package turn

import (
	"fmt"
	"os"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	clipboard     bool
	preserveOrder bool
}

// $ character named -v/ letter\ turned
// `LATIN CAPITAL LETTER TURNED H` is lower-case turned, in my fontl `K` and `T` are blocks
// pick appropriate near-matches otherwise, sometimes being the same letter
// # character browse -b IPA

const replacementsA string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const replacementsB string = "ɐqɔpə ᵷɥᴉ ʞꞁɯuodbɹsʇnʌʍxʎzⱯ      HI  ꞀƜ O   S  Ʌ X Z"

var turnCmd = &cobra.Command{
	Use:   "turn [text ...]",
	Short: "shows text upside-down",
	Run: func(cmd *cobra.Command, args []string) {
		if utf8.RuneCountInString(replacementsA) != utf8.RuneCountInString(replacementsB) {
			panic("hard error in source mismatched lengths of string replacements")
		}
		turnTable := make(map[rune]rune, 2*len(replacementsA))
		var ri, rj rune
		for i, j, wi, wj := 0, 0, 0, 0; i < len(replacementsA); {
			ri, wi = utf8.DecodeRuneInString(replacementsA[i:])
			rj, wj = utf8.DecodeRuneInString(replacementsB[j:])
			turnTable[ri] = rj
			turnTable[rj] = ri
			i += wi
			j += wj
		}

		text := make([]rune, 0, 2000)
		for argCount, arg := range args {
			if argCount > 0 {
				text = append(text, ' ')
			}
			for _, r := range arg {
				if t, found := turnTable[r]; found {
					text = append(text, t)
				} else {
					text = append(text, r)
				}
			}
		}

		if !flags.preserveOrder && len(text) >= 2 {
			last := len(text) - 1
			for i := 0; i <= (last+1)/2; i++ {
				text[i], text[last-i] = text[last-i], text[i]
			}
		}
		if len(text) > 0 && text[len(text)-1] != '\n' {
			text = append(text, '\n')
		}

		result := string(text)
		fmt.Print(result)

		if flags.clipboard {
			err := clipboard.WriteAll(result)
			if err != nil {
				root.Errored()
				fmt.Fprintf(os.Stderr, "clipboard write failure: %s\n", err)
			}
		}
	},
}

func init() {
	turnCmd.Flags().BoolVarP(&flags.clipboard, "clipboard", "c", false, "copy resulting chars to clipboard too")
	turnCmd.Flags().BoolVarP(&flags.preserveOrder, "preserve-order", "p", false, "keep characters in original order")
	if clipboard.Unsupported {
		// We don't want to only register the flag if clipboard is supported,
		// because that makes client portability more problematic.  Instead, we
		// just hide it to avoid offering something we can't honour, even
		// though we'll accept the option (and show an error) if given.
		turnCmd.Flags().MarkHidden("clipboard")
	}
	root.AddCommand(turnCmd)
}
