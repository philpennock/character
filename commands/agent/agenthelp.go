// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package agent

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/philpennock/character/commands/root"
	"github.com/philpennock/character/commands/version"
)

// AgentFlag describes one flag on an agent-visible command.
type AgentFlag struct {
	Name        string `json:"name"`
	Short       string `json:"short,omitempty"`
	Type        string `json:"type"`
	Default     string `json:"default"`
	Description string `json:"description"`
}

// AgentCommand describes one command (or sub-command) in the agent help JSON.
type AgentCommand struct {
	Name        string         `json:"name"`
	Usage       string         `json:"usage"`
	Short       string         `json:"short"`
	Flags       []AgentFlag    `json:"flags,omitempty"`
	SubCommands []AgentCommand `json:"subcommands,omitempty"`
}

// AgentHelp is the top-level object emitted by `agent help`.
type AgentHelp struct {
	Tool        string         `json:"tool"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Commands    []AgentCommand `json:"commands"`
}

var agentHelpCmd = &cobra.Command{
	Use:   "help",
	Short: "emit JSON schema of all commands and flags (agent-oriented)",
	Run:   runAgentHelp,
}

func init() {
	agentCmd.AddCommand(agentHelpCmd)
}

func runAgentHelp(cmd *cobra.Command, args []string) {
	rootCmd := root.Cobra()
	help := AgentHelp{
		Tool:        root.StaticProgramName,
		Version:     version.VersionString,
		Description: rootCmd.Short,
		Commands:    buildCommands(rootCmd.Commands()),
	}
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(help); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "agent help: marshal: %v\n", err)
		root.Errored()
	}
}

// buildCommands converts a slice of cobra commands into AgentCommand entries,
// excluding the "agent" command and its children (and hidden commands).
func buildCommands(cmds []*cobra.Command) []AgentCommand {
	result := make([]AgentCommand, 0, len(cmds))
	for _, cmd := range cmds {
		if cmd.Hidden {
			continue
		}
		if cmd.Name() == "agent" {
			continue
		}
		ac := AgentCommand{
			Name:  cmd.Name(),
			Usage: cmd.UseLine(),
			Short: cmd.Short,
		}

		// Collect non-hidden flags: local + inherited (for persistent flags
		// like --json that are registered on parent commands).
		seen := make(map[string]bool)
		collectFlags := func(fs *pflag.FlagSet) {
			fs.VisitAll(func(f *pflag.Flag) {
				if f.Hidden || seen[f.Name] {
					return
				}
				seen[f.Name] = true
				af := AgentFlag{
					Name:        f.Name,
					Type:        f.Value.Type(),
					Default:     f.DefValue,
					Description: f.Usage,
				}
				if f.Shorthand != "" {
					af.Short = f.Shorthand
				}
				ac.Flags = append(ac.Flags, af)
			})
		}
		collectFlags(cmd.Flags())
		collectFlags(cmd.InheritedFlags())

		if len(cmd.Commands()) > 0 {
			ac.SubCommands = buildCommands(cmd.Commands())
		}

		result = append(result, ac)
	}
	return result
}
