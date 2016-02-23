// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package transform

// $ character named -v/ letter\ turned
// `LATIN CAPITAL LETTER TURNED H` is lower-case turned, in my fontl `K` and `T` are blocks
// pick appropriate near-matches otherwise, sometimes being the same letter
// # character browse -b IPA

const turnReplacementsA string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const turnReplacementsB string = "ɐqɔpə ᵷɥᴉ ʞꞁɯuodbɹsʇnʌʍxʎzⱯ      HI  ꞀƜ O   S  Ʌ X Z"

var turnTable map[rune]rune

func transformTurn(args []string) (string, error) {
	ensureBidirectionalRuneTable(&turnTable, turnReplacementsA, turnReplacementsB)
	var transFlags transformFlags
	if !flags.preserveOrder {
		transFlags |= TRANS_REVERSE
	}
	return transformText(turnTable, args, transFlags), nil
}

var turnSubcommand = transformCobraCommand{
	Use:         "turn [text ...]",
	Short:       "turn characters upside down",
	Transformer: transformTurn,
}
