// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package puny

// This one is based on my older Python `puny` script.

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/net/idna"

	"github.com/philpennock/character/aux"
	"github.com/philpennock/character/encodings"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	encoding string
	hexInput bool
	punyIn   bool
	punyOut  bool
}

var punyCmd = &cobra.Command{
	Use:   "x-puny [string... [string...]]",
	Short: "handle punycode encoding/decoding of character sequences",
	Run: func(cmd *cobra.Command, args []string) {
		if flags.punyIn && flags.punyOut {
			root.Errorf("mutually incompatible flags --punycode-input and --punycode-output\n")
			return
		}

		// We don't use results tables at this time. Perhaps we should?
		// But we default to interpreting each item of input multiple ways.
		// Also our table setup is really geared towards individual unicode
		// characters, not punycode _sequences_.
		//
		// Not sure about the user-visible API here at all.
		// Nor if puny should be here.
		// Thus x-puny

		decoder, err := encodings.LoadCharsetDecoder(flags.encoding)
		if err != nil {
			root.Errorf("unable to get charset decoder: %s\n", err)
			return
		}

		// We first handle hex encoding, as being the most likely source of
		// non-UTF8 in UTF8 environments.
		if flags.hexInput {
			var errList []error
			args, errList = aux.HexDecodeArgs(args)
			for _, e := range errList {
				root.Errorf("error decoding hex: %s\n", e)
			}
		}

		for i, arg := range args {
			argUTF8, err := decoder.String(arg)
			if err != nil {
				root.Errorf("error decoding %q: %s\n", arg, err)
				continue
			}

			if i > 0 && !(flags.punyOut || flags.punyIn) {
				fmt.Println()
			}

			var (
				fromPunyToUnicode, fromUnicodeToPuny string
				errFPTU, errFUTP                     error
			)

			argPrefix := ""
			for _, p := range []string{"http://", "https://"} {
				if strings.HasPrefix(argUTF8, p) {
					argPrefix = p
					argUTF8 = argUTF8[len(p):]
				}
			}

			if !flags.punyOut {
				fromPunyToUnicode, errFPTU = idna.ToUnicode(argUTF8)
				if flags.punyIn {
					if errFPTU != nil {
						root.Errorf("error decoding arg %d %q from punycode: %s\n", i, arg, errFPTU)
					} else {
						fmt.Println(argPrefix + fromPunyToUnicode)
					}
					continue
				}
			}

			if !flags.punyIn {
				fromUnicodeToPuny, errFUTP = idna.ToASCII(argUTF8)
				if flags.punyOut {
					if errFUTP != nil {
						root.Errorf("error encoding arg %d %q to punycode: %s\n", i, arg, errFUTP)
					} else {
						fmt.Println(argPrefix + fromUnicodeToPuny)
					}
					continue
				}
			}

			fmt.Printf("Input: %s\nUnicode: %s%s\nPunycode: %s%s\n",
				/* no prefix, wasn't stripped from arg */ arg,
				argPrefix, fromPunyToUnicode,
				argPrefix, fromUnicodeToPuny)
			// not root.Errorf, see what failures we actually get in practice
			if errFPTU != nil {
				cmd.Printf("error decoding arg %d %q from punycode: %s\n", i, arg, errFPTU)
			}
			if errFUTP != nil {
				cmd.Printf("error encoding arg %d %q to punycode: %s\n", i, arg, errFUTP)
			}
		}
	},
}

func init() {
	punyCmd.Flags().StringVarP(&flags.encoding, "encoding", "e", "", "translate input from this encoding")
	punyCmd.Flags().BoolVarP(&flags.hexInput, "hex-input", "H", false, "take Hex-encoded input")
	punyCmd.Flags().BoolVarP(&flags.punyIn, "punycode-input", "i", false, "expect punycode, show only decode")
	punyCmd.Flags().BoolVarP(&flags.punyOut, "punycode-output", "o", false, "expect non-punycode, show only encode")

	root.AddCommand(punyCmd)
}
