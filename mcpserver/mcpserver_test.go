// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

func echoHandler(_ context.Context, args json.RawMessage) (string, error) {
	return string(args), nil
}

func failHandler(_ context.Context, _ json.RawMessage) (string, error) {
	return "", fmt.Errorf("intentional failure")
}

func TestNewServerAccessors(t *testing.T) {
	s := NewServer("test", "v0.1.0")
	if s.Name() != "test" {
		t.Errorf("Name() = %q, want %q", s.Name(), "test")
	}
	if s.Version() != "v0.1.0" {
		t.Errorf("Version() = %q, want %q", s.Version(), "v0.1.0")
	}
	if s.Instructions() != "" {
		t.Errorf("Instructions() = %q, want empty", s.Instructions())
	}

	s.SetInstructions("do things")
	if s.Instructions() != "do things" {
		t.Errorf("Instructions() = %q, want %q", s.Instructions(), "do things")
	}
}

func TestToolsEmpty(t *testing.T) {
	s := NewServer("test", "v1")
	if tools := s.Tools(); len(tools) != 0 {
		t.Errorf("Tools() returned %d entries on empty server, want 0", len(tools))
	}
}

func TestAddToolAndTools(t *testing.T) {
	s := NewServer("test", "v1")
	s.AddTool(ToolDef{
		Name:        "echo",
		Description: "echoes arguments",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}, echoHandler)
	s.AddTool(ToolDef{
		Name:        "fail",
		Description: "always fails",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}, failHandler)

	tools := s.Tools()
	if len(tools) != 2 {
		t.Fatalf("Tools() returned %d entries, want 2", len(tools))
	}

	// Registration order preserved.
	if tools[0].Name != "echo" {
		t.Errorf("tools[0].Name = %q, want %q", tools[0].Name, "echo")
	}
	if tools[1].Name != "fail" {
		t.Errorf("tools[1].Name = %q, want %q", tools[1].Name, "fail")
	}

	// Handlers are callable through the bridge type.
	result, err := tools[0].Handler(context.Background(), json.RawMessage(`{"x":1}`))
	if err != nil {
		t.Fatalf("echo handler error: %v", err)
	}
	if result != `{"x":1}` {
		t.Errorf("echo handler result = %q, want %q", result, `{"x":1}`)
	}

	_, err = tools[1].Handler(context.Background(), nil)
	if err == nil {
		t.Fatal("fail handler should return error")
	}
}

func TestToolsReturnsCopy(t *testing.T) {
	s := NewServer("test", "v1")
	s.AddTool(ToolDef{Name: "a", Description: "a"}, echoHandler)

	t1 := s.Tools()
	t2 := s.Tools()
	t1[0].Name = "mutated"
	if t2[0].Name != "a" {
		t.Error("Tools() returned shared backing, not independent copies")
	}
}

func TestToolRegistrationFields(t *testing.T) {
	s := NewServer("test", "v1")
	schema := json.RawMessage(`{"type":"object","properties":{"n":{"type":"string"}}}`)
	s.AddTool(ToolDef{
		Name:        "greet",
		Description: "says hello",
		InputSchema: schema,
	}, echoHandler)

	tr := s.Tools()[0]
	if tr.Name != "greet" {
		t.Errorf("Name = %q", tr.Name)
	}
	if tr.Description != "says hello" {
		t.Errorf("Description = %q", tr.Description)
	}
	if string(tr.InputSchema) != string(schema) {
		t.Errorf("InputSchema = %s", tr.InputSchema)
	}
	if tr.Handler == nil {
		t.Error("Handler is nil")
	}
}

func TestInnerNotNil(t *testing.T) {
	s := NewServer("test", "v1")
	if s.Inner() == nil {
		t.Error("Inner() returned nil")
	}
}
