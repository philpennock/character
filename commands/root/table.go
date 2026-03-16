// Copyright © 2015-2017,2025,2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package root

import (
	"os"

	"github.com/philpennock/character/internal/table"
)

func init() {
	// flagSet: match the logic from root.go
	flagSet := characterCmd.PersistentFlags()
	if table.MarkdownStyle != "" && isInsideAIAgent() {
		table.RenderStyle = table.MarkdownStyle
	}
	if table.AvailableStyles != nil {
		flagSet.StringVar(&table.RenderStyle, "table-style", table.RenderStyle, "style option for table")
	}
}

func tableRequestedList() bool {
	switch table.RenderStyle {
	case "", "?", "help":
		return true
	default:
		return false
	}
}

func isInsideAIAgent() bool {
	for _, v := range []string{
		"AGENT",          // multiple tools, emerging standard, name of agent as value, 'Goose', 'Amp'; <https://github.com/agentsmd/agents.md/issues/136>
		"AI_AGENT",       // promoted by Vercel's detect-agent package
		"OPENCODE_AGENT", // Open source multi-model client; set to '1'
		"CLAUDECODE",     // Anthropic's Claude Code, set to '1'
		"CURSOR_AGENT",   // OpenAI's Cursor, set to '1'
		"GEMINI_CLI",     // Google's Gemini, set to '1'
	} {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}
