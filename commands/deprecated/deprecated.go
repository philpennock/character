// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package deprecated

import (
	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

var frakturCmd = &cobra.Command{
	Use:    "fraktur",
	Short:  "toggle characters between plain & Fraktur",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		newArgs := make([]string, 0, 2+len(args))
		newArgs = append(newArgs, "transform", "fraktur")
		newArgs = append(newArgs, args...)
		root.Cobra().SetArgs(newArgs)
		err := root.Cobra().Execute()
		if err != nil {
			root.Errored()
			cmd.Printf("finding new fraktur cmd failed: %s\n", err.Error())
		}
	},

	// When ready to start spewing warnings:
	// Deprecated: "use 'transform fraktur'",
}

func init() {
	root.AddCommand(frakturCmd)
}
