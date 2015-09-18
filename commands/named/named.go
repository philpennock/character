package named

// FIXME: use cmd.Out() ?  what about non-error?

import (
	"fmt"
	"os"

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

var namedCmd = &cobra.Command{
	Use:   "named [name of character]",
	Short: "shows character with given name",
	Run: func(cmd *cobra.Command, args []string) {
		u := unicode.Load()
		var t *table.Table
		var errTable *table.Table
		if flags.verbose {
			t = table.New()
			t.AddHeaders(detailsHeaders()...)
			errTable = table.New()
		}

		for _, arg := range args {
			c, err := findChar(u, arg)
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
			t.AddRow(detailsFor(c)...)
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

func findChar(u unicode.Unicode, needle string) (rune, error) {
	// FIXME
	return 'c', nil
}

func detailsHeaders() []interface{} {
	return []interface{}{
		"Rune",
		"Name",
		"Foo",
		"Bar",
	}
}

func detailsFor(c rune) []interface{} {
	return []interface{}{
		"c",
		"dummy character c",
		"some metadata",
		"oops",
	}
}
