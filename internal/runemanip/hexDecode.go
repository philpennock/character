// Copyright © 2017,2021 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package runemanip

import (
	"bytes"
	"encoding/hex"
	"regexp"
	"strings"
	"unicode/utf16"
)

var matchHexSeq = regexp.MustCompile(`^(?:(?:[%=][0-9A-Fa-f]{2})|(?:\\u[0-9A-Fa-f]{4}))+`)

type malformedHexSequence struct {
	escapeSeq string
}

func (m malformedHexSequence) Error() string {
	return "malformed " + m.escapeSeq + "hex sequence"
}

// HexDecodeArgs decodes hex strings passed in argv, returning new args.
func HexDecodeArgs(in []string) (out []string, errList []error) {
	out = make([]string, len(in))

	// If we have %-encoded, then take non-%-preceded entries as literal character
	// If we don't, then assume that we just have hex strings
	// Handle = too, for MIME HDR encoding
	// Handle \u with _four_ digits too
	// We handle any, but only one per command-invocation.
	var escapeSeq string
	var escapeSeqBytes []byte
	needUTF16Decode := false

	for argN, arg := range in {
		if escapeSeq == "" {
			if strings.ContainsRune(arg, '%') {
				escapeSeq = "%"
			} else if strings.ContainsRune(arg, '=') {
				escapeSeq = "="
			} else if strings.Contains(arg, "\\u") {
				escapeSeq = "\\u"
				needUTF16Decode = true
			}
			if escapeSeq != "" {
				escapeSeqBytes = []byte(escapeSeq)
			}
		}

		if escapeSeq == "" || !strings.Contains(arg, escapeSeq) {
			out[argN] = in[argN]
			continue
		}

		argB := []byte(arg)
		chunks := make([][]byte, 0, len(argB))

		for len(argB) > 0 {
			nextEscape := bytes.Index(argB, escapeSeqBytes)
			if nextEscape < 0 {
				chunks = append(chunks, argB)
				argB = []byte{}
				continue
			}

			if nextEscape > 0 {
				chunks = append(chunks, argB[:nextEscape])
				argB = argB[nextEscape:]
			}

			matches := matchHexSeq.FindSubmatch(argB)
			if matches == nil {
				errList = append(errList, malformedHexSequence{escapeSeq})
				chunks = append(chunks, argB)
				argB = []byte{}
				continue
			}

			got := matches[0]
			argB = argB[len(got):]
			got = bytes.Replace(got, escapeSeqBytes, []byte{}, -1)
			// have an even-length sequence, length at least 2
			target := make([]byte, len(got)/2)
			n, err := hex.Decode(target, got)
			if err != nil {
				errList = append(errList, err)
				continue
			}
			if n != len(got)/2 {
				panic("oops, not got the right length buffer")
			}
			if needUTF16Decode {
				// we know from the regexp that for the \u case there are a multiple of 4 hexdigits
				sixteens := make([]uint16, n/2)
				for i := range sixteens {
					sixteens[i] = uint16(target[i*2])*256 + uint16(target[i*2+1])
				}
				runes := utf16.Decode(sixteens)
				target = []byte(string(runes))
			}
			chunks = append(chunks, target)
		}

		out[argN] = string(bytes.Join(chunks, []byte{}))
	}

	return
}
