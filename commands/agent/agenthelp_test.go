// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package agent_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/philpennock/character/commands/agent"
	"github.com/philpennock/character/commands/root"
	// Side-effect imports to populate the root command with subcommands.
	_ "github.com/philpennock/character/commands/named"
	_ "github.com/philpennock/character/commands/version"
)

func runAgentSubcmd(t *testing.T, args ...string) string {
	t.Helper()
	cmd := root.Cobra()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs(append([]string{"agent"}, args...))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute(%v): %v", args, err)
	}
	return buf.String()
}

func TestAgentHelpOutput(t *testing.T) {
	out := runAgentSubcmd(t, "help")

	var help agent.AgentHelp
	if err := json.Unmarshal([]byte(out), &help); err != nil {
		t.Fatalf("unmarshal JSON: %v\noutput:\n%s", err, out)
	}

	if help.Tool != "character" {
		t.Errorf("Tool = %q; want %q", help.Tool, "character")
	}
	if len(help.Commands) == 0 {
		t.Error("Commands is empty; expected at least one command")
	}

	// Find the "named" command.
	var namedCmd *agent.AgentCommand
	for i := range help.Commands {
		if help.Commands[i].Name == "named" {
			namedCmd = &help.Commands[i]
			break
		}
	}
	if namedCmd == nil {
		t.Fatal("no 'named' command in output")
	}

	// Verify --json flag appears.
	var hasJSON bool
	for _, f := range namedCmd.Flags {
		if f.Name == "json" && f.Type == "bool" {
			hasJSON = true
			break
		}
	}
	if !hasJSON {
		t.Errorf("named command flags = %v; want entry with Name=json Type=bool", namedCmd.Flags)
	}

	// Verify "agent" is not in the top-level commands.
	for _, c := range help.Commands {
		if c.Name == "agent" {
			t.Error("'agent' command should not appear in agent help output")
		}
	}
}
