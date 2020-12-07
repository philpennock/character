// Copyright © 2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package aliases

import (
	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
	"github.com/philpennock/character/resultset"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "alias for 'named -v/'",
	Run: func(cmd *cobra.Command, args []string) {
		newArgs := make([]string, 3, 3+len(args))
		newArgs[0] = "named"
		newArgs[1] = "--internal-implicit-verbose"
		newArgs[2] = "-/"
		newArgs = append(newArgs, args...)
		root.Cobra().SetArgs(newArgs)
		err := root.Cobra().Execute()
		if err != nil {
			root.Errored()
			cmd.Printf("finding 'named' command failed: %s\n", err.Error())
		}
	},
}

func init() {
	root.AddCommand(searchCmd)
	resultset.RegisterCmdFlags(searchCmd, true)
}
