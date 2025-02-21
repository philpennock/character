// Copyright © 2015-2017,2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package named

import (
	"errors"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/entities"
	"github.com/philpennock/character/internal/clipboard"
	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	clipboard  bool
	join       bool
	livevim    bool
	search     bool
	unsorted   bool
	htmlEntity bool
	xmlEntity  bool
}

// FIXME: make dedicated type, embed search info

// ErrUnknownCharacterName means the specified character does not exist
var ErrUnknownCharacterName = errors.New("unknown character name")

// ErrNoSearchResults means you're unlucky
var ErrNoSearchResults = errors.New("no search results")

type ErrorUnknownEntity string

func (e ErrorUnknownEntity) Error() string {
	return "unknown entity " + strconv.Quote(string(e))
}

type ErrorEntityNotKnownToUnicode rune

func (e ErrorEntityNotKnownToUnicode) Error() string {
	return "entity rune not in our Unicode tables, is 0x" + strconv.FormatInt(int64(e), 16)
}

var namedCmd = &cobra.Command{
	Use:   "named [name of character]",
	Short: "shows character with given name",
	Run: func(cmd *cobra.Command, args []string) {
		if err := resultset.FlagsOkay(); err != nil {
			root.Errorf("%s\n", err)
			return
		}
		srcs := sources.NewFast()
		if flags.search {
			srcs.LoadUnicodeSearch()
		}
		if flags.livevim {
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
				// The results are not sorted; we have []interface{}
				// To sort, we need to either create a way to have a results
				// set have sorting over sub-ranges, or work with a temporary
				// buffer.  Let's assume that for most cases, the number of
				// characters is small and sorting with temporary buffers is
				// sane, and provide an `unsorted` option for the caller for
				// the expected-to-be-unusual cases.
				if flags.unsorted {
					for _, cii := range found {
						results.AddCharInfo(cii.(unicode.CharInfo))
					}
				} else {
					tempCharInfoList := make(unicode.CharInfoList, len(found))
					for index, cii := range found {
						tempCharInfoList[index] = cii.(unicode.CharInfo)
					}
					tempCharInfoList.Sort()
					for _, ci := range tempCharInfoList {
						results.AddCharInfo(ci)
					}
				}
				continue
			}
			var entityMap map[string]rune
			if flags.xmlEntity {
				entityMap = entities.XMLEntities
			} else if flags.htmlEntity {
				entityMap = entities.HTMLEntities
			}
			if entityMap != nil {
				// XML entity names are case-sensitive
				ciRune, ok := entityMap[arg]
				if !ok {
					root.Errored()
					results.AddError(arg, ErrorUnknownEntity(arg))
					continue
				}
				ci, ok := srcs.Unicode.ByRune[ciRune]
				if !ok {
					root.Errored()
					results.AddError(arg, ErrorEntityNotKnownToUnicode(ciRune))
					continue
				}
				results.AddCharInfo(ci)
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
			err := clipboard.WriteAll(results.StringPlain(results.RunePrintType()))
			if err != nil {
				root.Errored()
				results.AddError("<clipboard>", err)
			}
		}

		results.RenderPerCmdline(resultset.PRINT_RUNE)
	},
}

func init() {
	namedCmd.Flags().BoolVarP(&flags.clipboard, "clipboard", "c", false, "copy resulting chars to clipboard too")
	namedCmd.Flags().BoolVarP(&flags.join, "join", "j", false, "all args are for one char name")
	namedCmd.Flags().BoolVarP(&flags.livevim, "livevim", "l", false, "load full vim data (for verbose)")
	namedCmd.Flags().BoolVarP(&flags.search, "search", "/", false, "search for words, not full name")
	namedCmd.Flags().BoolVarP(&flags.unsorted, "unsorted", "u", false, "do not sort search results")
	namedCmd.Flags().BoolVarP(&flags.htmlEntity, "html-entity", "H", false, "input words are HTML entity names")
	namedCmd.Flags().BoolVarP(&flags.xmlEntity, "xml-entity", "X", false, "input words are XML entity names")
	resultset.RegisterCmdFlags(namedCmd, true) // verbose v | net-verbose N | internal-debug; enable oneline
	if clipboard.Unsupported {
		// We don't want to only register the flag if clipboard is supported,
		// because that makes client portability more problematic.  Instead, we
		// just hide it to avoid offering something we can't honour, even
		// though we'll accept the option (and show an error) if given.
		namedCmd.Flags().MarkHidden("clipboard")
	}
	namedCmd.MarkFlagsMutuallyExclusive("search", "html-entity", "xml-entity")
	// FIXME: support verbose results without tables
	root.AddCommand(namedCmd)
}

func findCharInfoByName(u unicode.Unicode, needle string) (unicode.CharInfo, error) {
	n := strings.ToUpper(needle)
	// FIXME: wait, linear scan for each name?  Seriously?
	for k := range u.ByName {
		if k == n {
			return u.ByName[k], nil
		}
	}

	return unicode.CharInfo{}, ErrUnknownCharacterName
}
