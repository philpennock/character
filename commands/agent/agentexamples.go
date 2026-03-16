// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package agent

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

// AgentExample describes one example invocation for agent consumption.
type AgentExample struct {
	Category    string `json:"category"`
	Description string `json:"description"`
	Command     string `json:"command"`
	OutputShape string `json:"output_shape"`
}

var agentExamplesCmd = &cobra.Command{
	Use:   "examples [category]",
	Short: "emit JSON array of example invocations, optionally filtered by category",
	Args:  cobra.MaximumNArgs(1),
	Run:   runAgentExamples,
}

func init() {
	agentCmd.AddCommand(agentExamplesCmd)
}

var allExamples = []AgentExample{
	// lookup
	{
		Category:    "lookup",
		Description: "Look up a character glyph to get its Unicode name and properties",
		Command:     "character name -J ✓",
		OutputShape: "json:characters[0].name",
	},
	{
		Category:    "lookup",
		Description: "Find a character by its exact Unicode name",
		Command:     "character named -Jj CHECK MARK",
		OutputShape: "json:characters[0].display",
	},
	{
		Category:    "lookup",
		Description: "Look up a character by hex codepoint",
		Command:     "character code -J U+2713",
		OutputShape: "json:characters[0].name",
	},

	// search
	{
		Category:    "search",
		Description: "Search for characters whose names contain a word",
		Command:     "character named -J/ snowman",
		OutputShape: "json:characters[]",
	},
	{
		Category:    "search",
		Description: "Verbose search with a table of results",
		Command:     "character search checkmark",
		OutputShape: "table:multi-column",
	},

	// emoji
	{
		Category:    "emoji",
		Description: "Get a country flag emoji by two-letter country code",
		Command:     "character region FR",
		OutputShape: "plain:one line containing the flag glyph",
	},
	{
		Category:    "emoji",
		Description: "Combine an emoji with a skin-tone modifier",
		Command:     "character named -1j 'RAISED HAND' 'EMOJI MODIFIER FITZPATRICK TYPE-4'",
		OutputShape: "plain:combined glyph on one line",
	},

	// encoding
	{
		Category:    "encoding",
		Description: "Get UTF-8 bytes and escape sequences for a codepoint",
		Command:     "character code -J U+1F600",
		OutputShape: "json:characters[0].utf8 and .jsonEscape",
	},
	{
		Category:    "encoding",
		Description: "Get new encoding fields (utf8_bytes, unicode_escaped, rust_escaped)",
		Command:     "character named -Jj SNOWMAN",
		OutputShape: "json:characters[0].utf8_bytes characters[0].unicode_escaped",
	},

	// transform
	{
		Category:    "transform",
		Description: "Convert text to Fraktur mathematical letters",
		Command:     "character transform fraktur 'Hello World'",
		OutputShape: "plain:fraktur text on one line",
	},
	{
		Category:    "transform",
		Description: "Turn text upside down",
		Command:     "character transform turn 'Hello'",
		OutputShape: "plain:turned text on one line",
	},

	// browse
	{
		Category:    "browse",
		Description: "List all characters in a Unicode block",
		Command:     "character browse -b Arrows",
		OutputShape: "table:multi-column character table",
	},
	{
		Category:    "browse",
		Description: "List all Unicode block names",
		Command:     "character known -b",
		OutputShape: "plain:one block name per line",
	},
}

func runAgentExamples(cmd *cobra.Command, args []string) {
	var examples []AgentExample
	if len(args) == 0 {
		examples = allExamples
	} else {
		category := args[0]
		for _, ex := range allExamples {
			if ex.Category == category {
				examples = append(examples, ex)
			}
		}
	}
	if examples == nil {
		examples = []AgentExample{} // ensure JSON [] not null
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	if err := enc.Encode(examples); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "agent examples: marshal: %v\n", err)
		root.Errored()
	}
}
