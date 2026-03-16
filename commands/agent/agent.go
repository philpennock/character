// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// Package agent provides the `character agent` sub-command tree.  All output
// from these commands is stable, machine-readable JSON without ANSI escapes or
// box-drawing characters, intended for consumption by AI agents and tooling.
package agent

import (
	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "agent-oriented sub-commands with stable machine-readable output",
}

func init() {
	root.AddCommand(agentCmd)
}
