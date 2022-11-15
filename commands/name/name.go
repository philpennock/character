// Copyright © 2015-2017,2021-2022 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package name

import (
	"fmt"

	"github.com/spf13/cobra"
	"golang.org/x/net/idna"
	"golang.org/x/text/unicode/norm"

	"github.com/philpennock/character/internal/encodings"
	"github.com/philpennock/character/internal/runemanip"
	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	encoding      string
	listEncodings bool
	livevim       bool
	punyIn        bool
	hexInput      bool
}

var nameCmd = &cobra.Command{
	Use:   "name [char... [char...]]",
	Short: "shows information about supplied characters",
	Run: func(cmd *cobra.Command, args []string) {
		if err := resultset.FlagsOkay(); err != nil {
			root.Errorf("%s\n", err)
			return
		}

		if flags.listEncodings {
			// This should be deprecated, now that we have `known` top-level command
			cmd.Printf("%s %s: these names (and some aliases) are known:\n", root.Cobra().Name(), cmd.Name())
			for _, enc := range encodings.ListKnownCharsets() {
				cmd.Printf("\t%q\n", enc)
			}
			return
		}

		decoder, err := encodings.LoadCharsetDecoder(flags.encoding)
		if err != nil {
			root.Errorf("unable to get charset decoder: %s\n", err)
			return
		}

		srcs := sources.NewFast()
		if flags.livevim {
			srcs.LoadLiveVim()
		}
		approxCharCount := 0
		for _, a := range args {
			approxCharCount += len(a) + 1
		}
		results := resultset.New(srcs, approxCharCount)

		var pairedCodepoint rune = 0
		var priorCodepoint rune = 0

		// We first handle hex encoding, as being the most likely source of
		// non-UTF8 in UTF8 environments.
		if flags.hexInput {
			var errList []error
			args, errList = runemanip.HexDecodeArgs(args)
			for _, e := range errList {
				results.AddError("", e)
			}
		}

		for i, arg := range args {
			argUTF8, err := decoder.String(arg)
			if err != nil {
				results.AddError(arg, err)
				continue
			}
			if i > 0 {
				results.AddDivider()
			}
			if flags.punyIn {
				if t, err := idna.ToUnicode(argUTF8); err == nil {
					argUTF8 = t
				}
			}
			pairedCodepoint = 0
			for _, r := range argUTF8 {
				convertRune(r, &pairedCodepoint, &priorCodepoint, srcs, results, 0)
			}
		}

		results.RenderPerCmdline(resultset.PRINT_NAME)
	},
}

func init() {
	nameCmd.Flags().StringVarP(&flags.encoding, "encoding", "e", "", "translate input from this encoding")
	nameCmd.Flags().BoolVarP(&flags.hexInput, "hex-input", "H", false, "take Hex-encoded input")
	nameCmd.Flags().BoolVarP(&flags.listEncodings, "list-encodings", "", false, "list -e encodings & exit")
	nameCmd.Flags().BoolVarP(&flags.punyIn, "punycode-input", "p", false, "decode punycode on cmdline")
	resultset.RegisterCmdFlags(nameCmd, false) // verbose v | net-verbose N | internal-debug; don't enable oneline
	if resultset.CanTable() {
		nameCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data (for verbose)")
	}
	// FIXME: support verbose results without tables

	root.AddCommand(nameCmd)
}

func convertRune(r rune, pairedCodepoint *rune, priorCodepoint *rune, srcs *sources.Sources, results *resultset.ResultSet, originalRune rune) {
	ci, ok := srcs.Unicode.ByRune[r]
	if !ok {
		if originalRune == 0 {
			rs := string(r)
			decomp := norm.NFD.String(rs)
			if decomp != rs {
				for _, r2 := range decomp {
					convertRune(r2, pairedCodepoint, priorCodepoint, srcs, results, r)
				}
				return
			}
		}
		root.Errored()
		// FIXME: proper error type
		results.AddError(string(r), fmt.Errorf("unknown codepoint %x", int(r)))
	}

	results.AddCharInfoDerivedFrom(ci, originalRune)
	// Ancillary extra data if warranted
	if runemanip.IsVariationSelector(ci.Number) {
		if *priorCodepoint != 0 {
			results.AddStringSequence(string(*priorCodepoint) + string(ci.Number))
		}
		*priorCodepoint = 0
	} else if runemanip.IsPairCode(ci.Number) {
		if *pairedCodepoint != 0 {
			if ci2, ok := unicode.PairCharInfo(*pairedCodepoint, ci.Number); ok {
				results.AddCharInfoDerivedFrom(ci2, originalRune)
			} else {
				results.AddError("", fmt.Errorf("unknown codepair %x-%x", *pairedCodepoint, ci.Number))
			}
			*pairedCodepoint = 0
		} else {
			*pairedCodepoint = ci.Number
		}
		*priorCodepoint = 0
	} else {
		*priorCodepoint = ci.Number
	}
}
