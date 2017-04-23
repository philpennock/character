// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package region

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	clipboard bool
	noSpace   bool
}

// ErrInvalidRegionLetter means not known as a Regional Indicator character
var ErrInvalidRegionLetter = errors.New("unknown regional indicator character")

// RegionalIndicatorBase is the first rune in the A-Z range of the Regional Indicators.
const RegionalIndicatorBase = 0x1F1E6

var regionCmd = &cobra.Command{
	Use:   "region <X><Y> [<X1><Y1> ...]",
	Short: "performs Regional Indicator lookups",
	Run: func(cmd *cobra.Command, args []string) {
		regions := make([]string, 0, len(args))

	ArgIter:
		for i, arg := range args {
			parts := make([]rune, 0, len(arg))
			for c, letter := range arg {
				if c > 0 && c%2 == 0 && !flags.noSpace {
					parts = append(parts, ' ')
				}

				if letter >= 'a' && letter <= 'z' {
					parts = append(parts, letter-'a'+RegionalIndicatorBase)
				} else if letter >= 'A' && letter <= 'Z' {
					parts = append(parts, letter-'A'+RegionalIndicatorBase)
				} else {
					root.Errorf("arg %d: rune '%c' not a letter for a Regional Indicator\n", i+1, letter)
					continue ArgIter
				}
			}
			regions = append(regions, string(parts))
		}

		if flags.clipboard {
			err := clipboard.WriteAll(strings.Join(regions, " "))
			if err != nil {
				root.Errorf("clipboard: %s", err)
			}
		}
		for _, r := range regions {
			fmt.Println(r)
		}
	},
}

func init() {
	regionCmd.Flags().BoolVarP(&flags.clipboard, "clipboard", "c", false, "copy resulting chars to clipboard too")
	regionCmd.Flags().BoolVarP(&flags.noSpace, "nospace", "n", false, "no space between multiple flags in one cmdline parameter")
	if clipboard.Unsupported {
		// We don't want to only register the flag if clipboard is supported,
		// because that makes client portability more problematic.  Instead, we
		// just hide it to avoid offering something we can't honour, even
		// though we'll accept the option (and show an error) if given.
		regionCmd.Flags().MarkHidden("clipboard")
	}
	root.AddCommand(regionCmd)
}
