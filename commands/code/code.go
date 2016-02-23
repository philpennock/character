// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package code

import (
	"errors"
	"strconv"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	clipboard bool
	livevim   bool
	verbose   bool
}

// FIXME: make dedicated type, embed search info

// ErrUnknownCodepoint means the specified codepoint is not assigned
var ErrUnknownCodepoint = errors.New("unknown character codepoint")

var codeCmd = &cobra.Command{
	Use:   "code [codepoint ...]",
	Short: "shows character with codepoint",
	Run: func(cmd *cobra.Command, args []string) {
		srcs := sources.NewFast()
		if flags.verbose && flags.livevim {
			srcs.LoadLiveVim()
		}
		results := resultset.New(srcs, len(args))

		for _, arg := range args {
			ci, err := findCharInfoByCodepoint(srcs.Unicode, arg)
			if err != nil {
				root.Errored()
				results.AddError(arg, err)
				continue
			}
			results.AddCharInfo(ci)
		}

		if flags.clipboard {
			err := clipboard.WriteAll(results.StringPlain(resultset.PRINT_RUNE))
			if err != nil {
				root.Errored()
				results.AddError("<clipboard>", err)
			}
		}

		if flags.verbose {
			results.PrintTables()
		} else {
			results.PrintPlain(resultset.PRINT_RUNE)
		}
	},
}

func init() {
	codeCmd.Flags().BoolVarP(&flags.clipboard, "clipboard", "c", false, "copy resulting chars to clipboard too")
	codeCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data (for verbose)")
	if resultset.CanTable() {
		codeCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about the character")
	}
	if clipboard.Unsupported {
		// We don't want to only register the flag if clipboard is supported,
		// because that makes client portability more problematic.  Instead, we
		// just hide it to avoid offering something we can't honour, even
		// though we'll accept the option (and show an error) if given.
		codeCmd.Flags().MarkHidden("clipboard")
	}
	// FIXME: support verbose results without tables
	root.AddCommand(codeCmd)
}

func findCharInfoByCodepoint(u unicode.Unicode, needle string) (unicode.CharInfo, error) {
	n32, err := strconv.ParseInt(needle, 0, 32)
	if err != nil {
		return unicode.CharInfo{}, err
	}
	n := rune(n32)

	if ci, ok := u.ByRune[n]; ok {
		return ci, nil
	}
	return unicode.CharInfo{}, ErrUnknownCodepoint
}
