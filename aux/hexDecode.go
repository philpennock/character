// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package aux

import (
	"bytes"
	"encoding/hex"
	"regexp"
	"strings"
)

var matchHexPairSeq = regexp.MustCompile(`^(?:[%=][0-9A-Fa-f]{2})+`)

type malformedHexSequence struct {
	escapeChar rune
}

func (m malformedHexSequence) Error() string {
	return "malformed " + string(m.escapeChar) + "hex sequence"
}

// HexDecodeArgs decodes hex strings passed in argv, returning new args.
func HexDecodeArgs(in []string) (out []string, errList []error) {
	out = make([]string, len(in))

	// If we have %-encoded, then take non-%-preceded entries as literal character
	// If we don't, then assume that we just have hex strings
	// Handle = too, for MIME HDR encoding
	// We handle either, but only one or the other per command-invocation.
	var escapeChar rune = 0
	var escapeByteSeq []byte

	for argN, arg := range in {
		if escapeChar == 0 {
			if strings.ContainsRune(arg, '%') {
				escapeChar = '%'
				escapeByteSeq = []byte{'%'}
			} else if strings.ContainsRune(arg, '=') {
				escapeChar = '='
				escapeByteSeq = []byte{'='}
			}
		}

		if escapeChar == 0 || !strings.ContainsRune(arg, escapeChar) {
			out[argN] = in[argN]
			continue
		}

		argB := []byte(arg)
		chunks := make([][]byte, 0, len(argB))

		for len(argB) > 0 {
			nextEscape := bytes.IndexByte(argB, escapeByteSeq[0])
			if nextEscape < 0 {
				chunks = append(chunks, argB)
				argB = []byte{}
				continue
			}

			if nextEscape > 0 {
				chunks = append(chunks, argB[:nextEscape])
				argB = argB[nextEscape:]
			}

			matches := matchHexPairSeq.FindSubmatch(argB)
			if matches == nil {
				errList = append(errList, malformedHexSequence{escapeChar})
				chunks = append(chunks, argB)
				argB = []byte{}
				continue
			}

			got := matches[0]
			argB = argB[len(got):]
			got = bytes.Replace(got, escapeByteSeq, []byte{}, -1)
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
			chunks = append(chunks, target)
		}

		out[argN] = string(bytes.Join(chunks, []byte{}))
	}

	return
}
