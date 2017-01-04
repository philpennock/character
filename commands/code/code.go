// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package code

import (
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	base       intconvBase
	clipboard  bool
	netVerbose bool
	livevim    bool
	utf8hex    bool
	verbose    bool
}

// FIXME: make dedicated type, embed search info

// ErrUnknownCodepoint means the specified codepoint is not assigned
var ErrUnknownCodepoint = errors.New("unknown character codepoint")

// When decoding, we may encounter problems
type deferredError struct {
	arg string
	err error
}

var codeCmd = &cobra.Command{
	Use:   "code [codepoint ...]",
	Short: "shows character with codepoint",
	Run: func(cmd *cobra.Command, args []string) {
		srcs := sources.NewFast()
		if flags.verbose && flags.livevim {
			srcs.LoadLiveVim()
		}

		// If given utf8hex flag, then we need to pre-munge the parameters
		deferredErrors := make([]deferredError, 0, 20)

		if flags.utf8hex {
			// WAG the capacity
			newargs := make([]string, 0, len(args)*4)
			matchHexPairSeq := regexp.MustCompile(`^(?:%[0-9A-Fa-f]{2})+`)

			for _, arg := range args {
				// If we have %-encoded, then take non-%-preceded entries as literal character
				// If we don't, then assume that we just have hex strings
				if strings.ContainsRune(arg, '%') {
					for len(arg) > 0 {
						nextPercent := strings.IndexByte(arg, '%')
						if nextPercent < 0 {
							for _, c := range arg {
								newargs = append(newargs, strconv.Itoa(int(c)))
							}
							arg = ""
							continue
						}
						if nextPercent > 0 {
							for _, c := range arg[:nextPercent] {
								newargs = append(newargs, strconv.Itoa(int(c)))
							}
							arg = arg[nextPercent:]
							// do not 'continue', handle inline immediately next
						}
						matches := matchHexPairSeq.FindStringSubmatch(arg)
						if matches == nil {
							deferredErrors = append(deferredErrors, deferredError{arg: arg, err: errors.New("malformed %hex seq")})
							arg = ""
							continue
						}
						got := matches[0]
						arg = arg[len(got):]
						got = strings.Replace(got, "%", "", -1)

						toAdd, defErr := codepointsFromHexString(got)
						if toAdd != nil {
							newargs = append(newargs, toAdd...)
						}
						if defErr != nil {
							deferredErrors = append(deferredErrors, *defErr)
							continue
						}
					}
				} else {
					toAdd, defErr := codepointsFromHexString(arg)
					if toAdd != nil {
						newargs = append(newargs, toAdd...)
					}
					if defErr != nil {
						deferredErrors = append(deferredErrors, *defErr)
						continue
					}
				}
			}
			args = newargs
		}

		results := resultset.New(srcs, len(args))
		for _, e := range deferredErrors {
			results.AddError(e.arg, e.err)
		}

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
		} else if flags.netVerbose {
			results.SelectFieldsNet()
			results.PrintTables()
		} else {
			results.PrintPlain(resultset.PRINT_RUNE)
		}
	},
}

func init() {
	codeCmd.Flags().BoolVarP(&flags.clipboard, "clipboard", "c", false, "copy resulting chars to clipboard too")
	codeCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data (for verbose)")
	codeCmd.Flags().BoolVarP(&flags.utf8hex, "utf8hex", "H", false, "take UTF-8 Hex-encoded codepoint")
	codeCmd.Flags().VarP(&flags.base, "base", "b", "numeric base for code-ponts [default: usual parse rules]")
	if resultset.CanTable() {
		codeCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about the character")
		codeCmd.Flags().BoolVarP(&flags.netVerbose, "net-verbose", "N", false, "show net-biased information (punycode, etc)")
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
	n32, err := strconv.ParseInt(needle, flags.base.Int(), 32)
	if err != nil {
		return unicode.CharInfo{}, err
	}
	n := rune(n32)

	if ci, ok := u.ByRune[n]; ok {
		return ci, nil
	}
	return unicode.CharInfo{}, ErrUnknownCodepoint
}

func codepointsFromHexString(hs string) ([]string, *deferredError) {
	seq, err := hex.DecodeString(hs)
	if err != nil {
		return nil, &deferredError{arg: hs, err: err}
	}
	res := make([]string, 0, len(seq))
	errorOffset := 0
	for len(seq) > 0 {
		r, sz := utf8.DecodeRune(seq)
		if r == utf8.RuneError && (sz == 0 || sz == 1) {
			return res, &deferredError{arg: hs[errorOffset*2:], err: errors.New("invalid rune")}
		}
		if sz == 0 {
			return res, &deferredError{arg: hs[errorOffset*2:], err: errors.New("API broken (rune size 0)")}
		}
		seq = seq[sz:]
		errorOffset += sz
		res = append(res, strconv.Itoa(int(r)))
	}
	return res, nil
}
