// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package transform

import (
	"fmt"
	"os"

	"github.com/atotto/clipboard"
	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

var flags struct {
	clipboard     bool
	preserveOrder bool
	target        string
}

type transformer func(args []string) (result string, err error)

type transformCobraCommand struct {
	Use         string
	Short       string
	Transformer transformer
}

var transformCmd = &cobra.Command{
	Use:   "transform TYPE [text ...]",
	Short: "transform characters back and forth",
}

func transformWrapper(cmd *cobra.Command, args []string, transformer transformer) {
	result, err := transformer(args)
	if err != nil {
		root.Errored()
		cmd.Printf("%s: %s\n", cmd.Name(), err.Error())
		return
	}
	fmt.Print(result)
	if len(result) > 0 && result[len(result)-1] != '\n' {
		fmt.Print("\n")
	}

	if flags.clipboard {
		err := clipboard.WriteAll(result)
		if err != nil {
			root.Errored()
			fmt.Fprintf(os.Stderr, "clipboard write failure: %s\n", err)
		}
	}
}

func init() {
	transformCmd.PersistentFlags().BoolVarP(&flags.clipboard, "clipboard", "c", false, "copy resulting chars to clipboard too")
	transformCmd.PersistentFlags().BoolVarP(&flags.preserveOrder, "preserve-order", "p", false, "keep characters in original order")
	transformCmd.PersistentFlags().StringVarP(&flags.target, "target", "t", "", "map characters to this type")
	if clipboard.Unsupported {
		// We don't want to only register the flag if clipboard is supported,
		// because that makes client portability more problematic.  Instead, we
		// just hide it to avoid offering something we can't honour, even
		// though we'll accept the option (and show an error) if given.
		transformCmd.Flags().MarkHidden("clipboard")
	}

	for _, subCommand := range []transformCobraCommand{
		turnSubcommand,
		frakturSubcommand,
		mathSubcommand,
	} {
		c := subCommand
		transformCmd.AddCommand(&cobra.Command{
			Use:   c.Use,
			Short: c.Short,
			Run: func(cmd *cobra.Command, args []string) {
				transformWrapper(cmd, args, c.Transformer)
			},
		})
	}

	root.AddCommand(transformCmd)
}
