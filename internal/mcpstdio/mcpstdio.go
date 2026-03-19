// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// Package mcpstdio implements a minimal MCP (Model Context Protocol) stdio
// server.  MCP over stdio is JSON-RPC 2.0 with newline-delimited framing:
// each message is a single JSON object terminated by "\n", and messages
// MUST NOT contain embedded newlines.
//
// For a tool-only server, only four methods are required:
//
//	initialize              → InitializeResult
//	notifications/initialized → (no response)
//	tools/list              → ListToolsResult
//	tools/call              → CallToolResult
package mcpstdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Handler processes a single tool call.  It receives the tool arguments as
// raw JSON and returns a result string, or an error if the call failed.
type Handler func(ctx context.Context, args json.RawMessage) (string, error)

// ToolDef describes a single MCP tool for registration and for tools/list.
type ToolDef struct {
	Name        string
	Description string
	InputSchema json.RawMessage // hand-written JSON Schema object
}

type toolEntry struct {
	def     ToolDef
	handler Handler
}

// Server is a tool-only MCP stdio server.
type Server struct {
	name         string
	version      string
	instructions string
	tools        []toolEntry
	byName       map[string]int
}

// NewServer creates a Server with the given name and version strings (used in
// the InitializeResult serverInfo).
func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
		byName:  make(map[string]int),
	}
}

// SetInstructions sets the instructions string returned in the InitializeResult.
// Per MCP 2025-03-26 §Lifecycle, clients MAY surface this to the model to guide
// tool discovery and usage.
func (s *Server) SetInstructions(text string) {
	s.instructions = text
}

// AddTool registers a tool.  Registration order is preserved in tools/list.
func (s *Server) AddTool(def ToolDef, h Handler) {
	s.byName[def.Name] = len(s.tools)
	s.tools = append(s.tools, toolEntry{def: def, handler: h})
}

// ServeStdio runs the MCP server on os.Stdin / os.Stdout.
func (s *Server) ServeStdio(ctx context.Context) error {
	return s.ServeConn(ctx, os.Stdin, os.Stdout)
}

// ServeConn runs the MCP server on the given reader/writer.  It is the
// testable entry point; production code calls ServeStdio.
func (s *Server) ServeConn(ctx context.Context, r io.Reader, w io.Writer) error {
	br := bufio.NewReader(r)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		body, err := readFrame(br)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("mcpstdio: read: %w", err)
		}

		var req struct {
			JSONRPC string          `json:"jsonrpc"`
			ID      json.RawMessage `json:"id"`
			Method  string          `json:"method"`
			Params  json.RawMessage `json:"params"`
		}
		if err := json.Unmarshal(body, &req); err != nil {
			writeError(w, json.RawMessage("null"), -32700, "Parse error")
			continue
		}

		switch req.Method {
		case "initialize":
			s.handleInitialize(w, req.ID)
		case "notifications/initialized":
			// Notification: no response required.
		case "tools/list":
			s.handleToolsList(w, req.ID)
		case "tools/call":
			s.handleToolsCall(ctx, w, req.ID, req.Params)
		default:
			// Only respond if this looks like a request (has an id).
			if len(req.ID) > 0 {
				writeError(w, req.ID, -32601, "Method not found")
			}
		}
	}
}

// readFrame reads one newline-terminated JSON-RPC message from r.
func readFrame(r *bufio.Reader) ([]byte, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	// Trim the trailing newline (and any \r before it).
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	return line, nil
}

// writeFrame writes one newline-terminated JSON message to w.
func writeFrame(w io.Writer, body []byte) {
	w.Write(body)   //nolint:errcheck
	fmt.Fprintln(w) //nolint:errcheck
}

func writeResponse(w io.Writer, id json.RawMessage, result any) {
	type response struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Result  any             `json:"result"`
	}
	body, _ := json.Marshal(response{JSONRPC: "2.0", ID: id, Result: result})
	writeFrame(w, body)
}

func writeError(w io.Writer, id json.RawMessage, code int, message string) {
	type rpcError struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	type response struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Error   rpcError        `json:"error"`
	}
	body, _ := json.Marshal(response{
		JSONRPC: "2.0",
		ID:      id,
		Error:   rpcError{Code: code, Message: message},
	})
	writeFrame(w, body)
}

func (s *Server) handleInitialize(w io.Writer, id json.RawMessage) {
	type serverInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	type capabilities struct {
		Tools map[string]any `json:"tools"`
	}
	type result struct {
		ProtocolVersion string       `json:"protocolVersion"`
		Capabilities    capabilities `json:"capabilities"`
		ServerInfo      serverInfo   `json:"serverInfo"`
		Instructions    string       `json:"instructions,omitempty"`
	}
	writeResponse(w, id, result{
		ProtocolVersion: "2024-11-05",
		Capabilities:    capabilities{Tools: map[string]any{}},
		ServerInfo:      serverInfo{Name: s.name, Version: s.version},
		Instructions:    s.instructions,
	})
}

func (s *Server) handleToolsList(w io.Writer, id json.RawMessage) {
	type toolItem struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		InputSchema json.RawMessage `json:"inputSchema"`
	}
	type result struct {
		Tools []toolItem `json:"tools"`
	}
	items := make([]toolItem, len(s.tools))
	for i, t := range s.tools {
		items[i] = toolItem{
			Name:        t.def.Name,
			Description: t.def.Description,
			InputSchema: t.def.InputSchema,
		}
	}
	writeResponse(w, id, result{Tools: items})
}

func (s *Server) handleToolsCall(ctx context.Context, w io.Writer, id json.RawMessage, params json.RawMessage) {
	var p struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal(params, &p); err != nil {
		writeError(w, id, -32602, "Invalid params")
		return
	}

	idx, ok := s.byName[p.Name]
	if !ok {
		writeError(w, id, -32601, fmt.Sprintf("Unknown tool: %s", p.Name))
		return
	}

	type contentItem struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	type callResult struct {
		Content []contentItem `json:"content"`
		IsError bool          `json:"isError,omitempty"`
	}

	text, err := s.tools[idx].handler(ctx, p.Arguments)
	if err != nil {
		writeResponse(w, id, callResult{
			Content: []contentItem{{Type: "text", Text: err.Error()}},
			IsError: true,
		})
		return
	}
	writeResponse(w, id, callResult{
		Content: []contentItem{{Type: "text", Text: text}},
	})
}
