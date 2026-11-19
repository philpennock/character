// Copyright © 2016,2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package transform

// $ character named -v/ letter\ turned
// `LATIN CAPITAL LETTER TURNED H` is lower-case turned, in my font; `K` and `T` are blocks
// pick appropriate near-matches otherwise, sometimes being the same letter
// # character browse -b IPA
// Updated 2026-11 and now 'B' and 'R' use Lisu/Fraser alphabet turns, which might have font issues.
// Also using a much better 'T'.

// <http://www.upsidedowntext.com/unicode> has more ☺

const turnReplacementsA string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789?!&()[]{}<>.,';\"_"
const turnReplacementsB string = "ɐqɔpǝɟᵷɥᴉɾʞꞁɯuodbɹsʇnʌʍxʎzⱯꓭƆᗡƎℲ⅁HIſꞰꞀƜNOԀ ꓤSꞱ∩ɅMX⅄Z0Ɩ↊↋ ϛ9 86¿¡⅋)(][}{><˙ʻ,⸵„‾"

var turnTable map[rune]rune

// TransformTurn converts each argument to upside-down characters and returns
// the results joined by spaces (text is reversed by default, matching the CLI
// default of preserveOrder=false).  It is the exported API for non-Cobra
// callers (e.g. the MCP server).
func TransformTurn(args []string) (string, error) {
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
	Transformer: TransformTurn,
}
