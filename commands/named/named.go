package named

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	search    bool
	substring bool
	verbose   bool
}

// FIXME: make dedicated type, embed search info
var ErrUnknownCharacterName = errors.New("unknown character name")

// XXX
// should move unicode into a DataSources, to handle vim/xhtml/html/etc
// should have a ResultsSet system, move table/display logic into that

var namedCmd = &cobra.Command{
	Use:   "named [name of character]",
	Short: "shows character with given name",
	Run: func(cmd *cobra.Command, args []string) {
		srcs := sources.NewAll()
		results := resultset.New(srcs, len(args))

		for _, arg := range args {
			ci, err := findCharInfoByName(srcs.Unicode, arg)
			if err != nil {
				root.Errored()
				results.AddError(arg, err)
				continue
			}
			results.AddCharInfo(ci)
		}

		if flags.verbose {
			results.PrintTables()
		} else {
			results.PrintPlain(resultset.PRINT_RUNE)
		}
	},
}

func init() {
	namedCmd.Flags().BoolVarP(&flags.search, "search", "/", false, "search for words, not full name")
	namedCmd.Flags().BoolVarP(&flags.substring, "substring", "s", false, "search for substrings, not words")
	namedCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about the character")
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
