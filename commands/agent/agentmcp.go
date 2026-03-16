// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package agent

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/agent/mcpserver"
	"github.com/philpennock/character/commands/root"
	"github.com/philpennock/character/sources"
)

var agentMCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "start an MCP stdio server exposing Unicode lookups as tools",
	Long: `Start an MCP (Model Context Protocol) server on stdin/stdout.

The server speaks JSON-RPC 2.0 with Content-Length framing and exposes
eight Unicode lookup tools.  The search index is loaded eagerly at startup
(~100-300ms) so that the first search request is fast.`,
	Run: runAgentMCP,
}

func init() {
	agentCmd.AddCommand(agentMCPCmd)
}

func runAgentMCP(cmd *cobra.Command, args []string) {
	srcs := sources.NewFast()
	srcs.LoadUnicodeSearch()

	srv := mcpserver.NewServer(srcs)
	if err := srv.ServeStdio(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "agent mcp: %v\n", err)
		root.Errored()
	}
}
