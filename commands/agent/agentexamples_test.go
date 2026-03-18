// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package agent_test

import (
	"encoding/json"
	"testing"

	"github.com/philpennock/character/commands/agent"
)

func TestAgentExamplesAll(t *testing.T) {
	out := runAgentSubcmd(t, "examples")

	var examples []agent.AgentExample
	if err := json.Unmarshal([]byte(out), &examples); err != nil {
		t.Fatalf("unmarshal JSON: %v\noutput:\n%s", err, out)
	}
	if len(examples) == 0 {
		t.Fatal("expected non-empty examples list")
	}

	categories := make(map[string]bool)
	for _, ex := range examples {
		categories[ex.Category] = true
	}
	for _, want := range []string{"lookup", "search", "emoji", "encoding", "transform", "browse"} {
		if !categories[want] {
			t.Errorf("category %q missing from examples", want)
		}
	}
}

func TestAgentExamplesFilteredCategory(t *testing.T) {
	out := runAgentSubcmd(t, "examples", "lookup")

	var examples []agent.AgentExample
	if err := json.Unmarshal([]byte(out), &examples); err != nil {
		t.Fatalf("unmarshal JSON: %v\noutput:\n%s", err, out)
	}
	if len(examples) == 0 {
		t.Fatal("expected non-empty lookup examples")
	}
	for _, ex := range examples {
		if ex.Category != "lookup" {
			t.Errorf("got example with Category=%q; want only %q", ex.Category, "lookup")
		}
	}
}

func TestAgentExamplesUnknownCategory(t *testing.T) {
	out := runAgentSubcmd(t, "examples", "nonexistent-category-xyz")

	var examples []agent.AgentExample
	if err := json.Unmarshal([]byte(out), &examples); err != nil {
		t.Fatalf("unmarshal JSON: %v\noutput:\n%s", err, out)
	}
	if examples == nil {
		t.Error("expected empty array [], not null")
	}
	if len(examples) != 0 {
		t.Errorf("expected 0 examples for unknown category, got %d", len(examples))
	}
}
