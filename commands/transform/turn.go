// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package transform

// $ character named -v/ letter\ turned
// `LATIN CAPITAL LETTER TURNED H` is lower-case turned, in my fontl `K` and `T` are blocks
// pick appropriate near-matches otherwise, sometimes being the same letter
// # character browse -b IPA

// <http://www.upsidedowntext.com/unicode> has more ☺

const turnReplacementsA string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789?!&()[]{}<>"
const turnReplacementsB string = "ɐqɔpəɟᵷɥᴉɾʞꞁɯuodbɹsʇnʌʍxʎzⱯ Ɔ ƎℲפHIſ ꞀƜNOԀ  S┴∩Ʌ X⅄Z0     9 86¿¡⅋)(][}{><"

var turnTable map[rune]rune

func transformTurn(args []string) (string, error) {
	ensureBidirectionalRuneTable(&turnTable, turnReplacementsA, turnReplacementsB)
	var transFlags transformFlags
	if !flags.preserveOrder {
		transFlags |= kTRANS_REVERSE
	}
	return transformText(turnTable, args, transFlags), nil
}

var turnSubcommand = transformCobraCommand{
	Use:         "turn [text ...]",
	Short:       "turn characters upside down",
	Transformer: transformTurn,
}
