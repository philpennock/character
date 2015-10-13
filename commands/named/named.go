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
	join      bool
	search    bool
	substring bool
	verbose   bool
}

// FIXME: make dedicated type, embed search info
var ErrUnknownCharacterName = errors.New("unknown character name")

var namedCmd = &cobra.Command{
	Use:   "named [name of character]",
	Short: "shows character with given name",
	Run: func(cmd *cobra.Command, args []string) {
		srcs := sources.NewAll()
		results := resultset.New(srcs, len(args))

		if flags.join {
			args = []string{strings.Join(args, " ")}
		}

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
	namedCmd.Flags().BoolVarP(&flags.join, "join", "j", false, "all args are for one char name")
	namedCmd.Flags().BoolVarP(&flags.search, "search", "/", false, "search for words, not full name")
	namedCmd.Flags().BoolVarP(&flags.substring, "substring", "s", false, "search for substrings, not words")
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
