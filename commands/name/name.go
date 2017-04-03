// Copyright © 2015-2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package name

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/net/idna"
	"golang.org/x/text/unicode/norm"

	"github.com/philpennock/character/metadata"
	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	internalDebug bool
	livevim       bool
	netVerbose    bool
	punyIn        bool
	verbose       bool
}

var nameCmd = &cobra.Command{
	Use:   "name [char... [char...]]",
	Short: "shows information about supplied characters",
	Run: func(cmd *cobra.Command, args []string) {
		srcs := sources.NewFast()
		if flags.verbose && flags.livevim {
			srcs.LoadLiveVim()
		}
		approxCharCount := 0
		for _, a := range args {
			approxCharCount += len(a) + 1
		}
		results := resultset.New(srcs, approxCharCount)

		var pairedCodepoint rune = 0

		for i, arg := range args {
			if i > 0 {
				results.AddDivider()
			}
			if flags.punyIn {
				if t, err := idna.ToUnicode(arg); err == nil {
					arg = t
				}
			}
			pairedCodepoint = 0
			for _, r := range arg {
				convertRune(r, &pairedCodepoint, srcs, results, false)
			}
		}

		if flags.verbose {
			results.PrintTables()
		} else if flags.netVerbose {
			results.SelectFieldsNet()
			results.PrintTables()
		} else if flags.internalDebug {
			results.SelectFieldsDebug()
			results.PrintTables()
		} else {
			results.PrintPlain(resultset.PRINT_NAME)
		}
	},
}

func init() {
	if resultset.CanTable() {
		nameCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data (for verbose)")
		nameCmd.Flags().BoolVarP(&flags.netVerbose, "net-verbose", "N", false, "show net-biased information (punycode, etc)")
		nameCmd.Flags().BoolVarP(&flags.internalDebug, "internal-debug", "", false, "")
		nameCmd.Flags().MarkHidden("internal-debug")
		nameCmd.Flags().BoolVarP(&flags.punyIn, "punycode-input", "p", false, "decode punycode on cmdline")
		nameCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about the character")
	}
	// FIXME: support verbose results without tables

	root.AddCommand(nameCmd)
}

func convertRune(r rune, pairedCodepoint *rune, srcs *sources.Sources, results *resultset.ResultSet, isNormalized bool) {
	ci, ok := srcs.Unicode.ByRune[r]
	if !ok {
		if !isNormalized {
			rs := string(r)
			decomp := norm.NFD.String(rs)
			if decomp != rs {
				for _, r2 := range decomp {
					convertRune(r2, pairedCodepoint, srcs, results, true)
				}
				return
			}
		}
		root.Errored()
		// FIXME: proper error type
		results.AddError(string(r), fmt.Errorf("unknown codepoint %x", int(r)))
	}

	results.AddCharInfo(ci)
	// Ancillary extra data if warranted
	if metadata.IsPairCode(ci.Number) {
		if *pairedCodepoint != 0 {
			if ci2, ok := metadata.PairCharInfo(*pairedCodepoint, ci.Number); ok {
				results.AddCharInfo(ci2)
			} else {
				results.AddError("", fmt.Errorf("unknown codepair %x-%x", *pairedCodepoint, ci.Number))
			}
			*pairedCodepoint = 0
		} else {
			*pairedCodepoint = ci.Number
		}
	}
}
