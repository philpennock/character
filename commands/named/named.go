package named

// FIXME: use cmd.Out() ?  what about non-error?

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/table"
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
type Sources struct {
	unicode unicode.Unicode
}

func Sources__New() Sources {
	return Sources{
		unicode: unicode.Load(),
	}
}

var namedCmd = &cobra.Command{
	Use:   "named [name of character]",
	Short: "shows character with given name",
	Run: func(cmd *cobra.Command, args []string) {
		sources := Sources__New()
		var t *table.Table
		var errTable *table.Table
		if flags.verbose {
			t = table.New()
			t.AddHeaders(detailsHeaders()...)
			errTable = table.New()
		}

		for _, arg := range args {
			c, err := sources.findCharByName(arg)
			if err != nil {
				root.Errored()
			}
			if !flags.verbose {
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %s\n", err)
				} else {
					fmt.Printf("%c\n", c)
				}
				continue
			}
			if err != nil {
				errTable.AddRow(arg, err)
				continue
			}
			t.AddRow(sources.detailsFor(c)...)
		}

		if !flags.verbose {
			return
		}
		if !t.Empty() {
			fmt.Print(t.Render())
		}
		if !errTable.Empty() {
			fmt.Fprint(os.Stderr, "Errors:\n")
			fmt.Fprint(os.Stderr, errTable.Render())
		}
	},
}

func init() {
	namedCmd.Flags().BoolVarP(&flags.search, "search", "/", false, "search for words, not full name")
	namedCmd.Flags().BoolVarP(&flags.substring, "substring", "s", false, "search for substrings, not words")
	namedCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show information about the character")
	root.AddCommand(namedCmd)
}

func (s Sources) findCharByName(needle string) (rune, error) {
	n := strings.ToUpper(needle)
	for k := range s.unicode.ByName {
		if k == n {
			return s.unicode.ByName[k].Number, nil
		}
	}

	return 0, ErrUnknownCharacterName
}

func detailsHeaders() []interface{} {
	return []interface{}{
		"Rune",
		"Name",
		"Foo",
		"Bar",
	}
}

func (s Sources) detailsFor(r rune) []interface{} {
	return []interface{}{
		string(r),
		s.unicode.ByRune[r].Name,
		"todo",
		"todo",
	}
}
