// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package named

import (
	"errors"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	clipboard bool
	join      bool
	livevim   bool
	search    bool
	verbose   bool
}

// FIXME: make dedicated type, embed search info

// ErrUnknownCharacterName means the specified character does not exist
var ErrUnknownCharacterName = errors.New("unknown character name")

// ErrNoSearchResults means you're unlucky
var ErrNoSearchResults = errors.New("no search results")

var namedCmd = &cobra.Command{
	Use:   "named [name of character]",
	Short: "shows character with given name",
	Run: func(cmd *cobra.Command, args []string) {
		srcs := sources.NewFast()
		if flags.search {
			srcs.LoadUnicodeSearch()
		}
		if flags.verbose && flags.livevim {
			srcs.LoadLiveVim()
		}
		results := resultset.New(srcs, len(args))

		if flags.join {
			args = []string{strings.Join(args, " ")}
		}

		for _, arg := range args {
			if flags.search {
				// works as of github.com/argusdusty/Ferret dd9e84e (2015-10-12)
				// prior to that, could work around limit with 0, but when author
				// made -1 work per contract, he made a 0 limit return 0 results
				_, found := srcs.Unicode.Search.Query(arg, -1)
				if len(found) == 0 {
					root.Errored()
					results.AddError(arg, ErrNoSearchResults)
					continue
				}
				for _, cii := range found {
					results.AddCharInfo(cii.(unicode.CharInfo))
				}
				continue
			}
			ci, err := findCharInfoByName(srcs.Unicode, arg)
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
	namedCmd.Flags().BoolVarP(&flags.clipboard, "clipboard", "c", false, "copy resulting chars to clipboard too")
	namedCmd.Flags().BoolVarP(&flags.join, "join", "j", false, "all args are for one char name")
	namedCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data (for verbose)")
	namedCmd.Flags().BoolVarP(&flags.search, "search", "/", false, "search for words, not full name")
	if resultset.CanTable() {
		namedCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about the character")
	}
	// FIXME: support verbose results without tables
	root.AddCommand(namedCmd)
}

func findCharInfoByName(u unicode.Unicode, needle string) (unicode.CharInfo, error) {
	n := strings.ToUpper(needle)
	for k := range u.ByName {
		if k == n {
			return u.ByName[k], nil
		}
	}

	return unicode.CharInfo{}, ErrUnknownCharacterName
}
