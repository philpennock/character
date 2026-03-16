// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// Package mcpstdio implements a minimal MCP (Model Context Protocol) stdio
// server.  MCP over stdio is JSON-RPC 2.0 with Content-Length framing.
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
	"strconv"
	"strings"
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
	name    string
	version string
	tools   []toolEntry
	byName  map[string]int
}

// NewServer creates a Server with the given name and version strings (used in
// the InitializeResult serverInfo).
func NewServer(name, version string) *Server {
	return &Server{
		name:   name,
		version: version,
		byName: make(map[string]int),
	}
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

// readFrame reads one Content-Length-framed JSON-RPC message from r.
func readFrame(r *bufio.Reader) ([]byte, error) {
	contentLength := -1
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if after, ok := strings.CutPrefix(line, "Content-Length:"); ok {
			val := strings.TrimSpace(after)
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("mcpstdio: invalid Content-Length: %q", val)
			}
			contentLength = n
		}
		// Other headers (e.g. Content-Type) are accepted and ignored.
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("mcpstdio: missing Content-Length header")
	}
	buf := make([]byte, contentLength)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("mcpstdio: read body: %w", err)
	}
	return buf, nil
}

// writeFrame writes one Content-Length-framed message to w.
func writeFrame(w io.Writer, body []byte) {
	fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(body))
	w.Write(body) //nolint:errcheck
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
	}
	writeResponse(w, id, result{
		ProtocolVersion: "2024-11-05",
		Capabilities:    capabilities{Tools: map[string]any{}},
		ServerInfo:      serverInfo{Name: s.name, Version: s.version},
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
