// Copyright © 2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package sources

import (
	"strconv"
	"strings"
)

// X11Data is the set of all data we have about X11 keyboard sequences.
// Note: we only have static compile-time data, we don't (at this time?) support ~/.XCompose
// TODO: support ~/.XCompose using the same liveness flag as for vim?  New flag?
type X11Data struct {
	// The string is a space-separated list of input sequences.
	// An ASCII SPACE 0x20 will not appear, as we encode it as `␠` 0x2420.
	DigraphsByRune map[rune]string
}

// DigraphsFor retrieves a string which is a space-separated list of the known
// digraph sequences which will produce a given rune.
// Note that we use `␠` in place of space in our representation.
func (x X11Data) DigraphsFor(r rune) string {
	if s, ok := x.DigraphsByRune[r]; ok {
		return s
	}
	return ""
}

// DigraphsSliceFor returns a list of input sequences which emit the given rune.
// We take firm advantage of our ␠ usage here.
func (x X11Data) DigraphsSliceFor(r rune) []string {
	if s, ok := x.DigraphsByRune[r]; ok {
		return strings.Split(s, " ")
	}
	return nil
}

func loadStaticX11Digraphs() X11Data {
	x := X11Data{
		DigraphsByRune: make(map[rune]string, staticX11RuneCount),
	}
	for _, l := range strings.Split(staticX11ComposeSeqs, "\n") {
		if len(l) == 0 {
			continue
		}
		parts := strings.SplitN(l, "\t", 2)
		ri, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err.Error())
		}
		x.DigraphsByRune[rune(ri)] = parts[1]
	}
	return x
}
