// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package fraktur

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

// We rely upon Unicode being unchanging in codepoint assignments.
// There are non-bold frakturs for all letters except these capitals: CHIRZ
// I mean, seriously, why?

const (
	frakturCapitalA     = 0x1d504
	frakturCapitalZish  = 0x1d51c
	frakturSmallA       = 0x1d51e
	frakturSmallZ       = 0x1d537
	frakturBoldCapitalA = 0x1d56c
	frakturBoldCapitalZ = 0x1d585
	frakturBoldSmallA   = 0x1d586
	frakturBoldSmallZ   = 0x1d59f
)

func toggleRune(r rune) rune {
	switch r {
	case 'C', 'H', 'I', 'R', 'Z':
		return r - 'A' + frakturBoldCapitalA
	}
	switch {
	case (r >= 'A') && (r <= 'Z'):
		return r - 'A' + frakturCapitalA
	case r >= 'a' && r <= 'z':
		return r - 'a' + frakturSmallA
	case r >= frakturCapitalA && r <= frakturCapitalZish: // MATHEMATICAL FRAKTUR CAPITAL
		// CHIRZ are not assigned, but there are gaps in the assignments left
		// for them, so if we get one of those code-points, map it anyway
		return r + 'A' - frakturCapitalA
	case r >= frakturSmallA && r <= frakturSmallZ: // MATHEMATICAL FRAKTUR SMALL
		return r + 'a' - frakturSmallA
	case r >= frakturBoldCapitalA && r <= frakturBoldCapitalZ: // MATHEMATICAL BOLD FRAKTUR CAPITAL
		return r + 'A' - frakturBoldCapitalA
	case r >= frakturBoldSmallA && r <= frakturBoldSmallZ: // MATHEMATICAL BOLD FRAKTUR SMALL
		return r + 'a' - frakturBoldSmallA
	}
	return r
}

var frakturCmd = &cobra.Command{
	Use:   "fraktur",
	Short: "toggle characters between plain & Fraktur",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			return
		}
		output := make([]string, len(args))
		for argI := range args {
			output[argI] = strings.Map(toggleRune, args[argI])
		}
		_, _ = fmt.Println(strings.Join(output, " "))
	},
}

func init() {
	root.AddCommand(frakturCmd)
}
