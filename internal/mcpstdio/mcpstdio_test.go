// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpstdio_test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/philpennock/character/internal/mcpstdio"
)

// testClient creates an in-process pipe-based client/server pair.
// send writes a framed JSON-RPC request and synchronously reads the next
// framed response.  close shuts down both pipes and waits for ServeConn.
func testClient(t *testing.T, srv *mcpstdio.Server) (
	send func(method string, params any) json.RawMessage,
	closeFunc func(),
) {
	t.Helper()

	// client writes → server reads
	serverR, clientW := io.Pipe()
	// server writes → client reads
	clientR, serverW := io.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- srv.ServeConn(ctx, serverR, serverW)
	}()

	clientBuf := bufio.NewReader(clientR)
	msgID := 0

	send = func(method string, params any) json.RawMessage {
		t.Helper()
		msgID++
		var paramsRaw json.RawMessage
		if params != nil {
			var err error
			paramsRaw, err = json.Marshal(params)
			if err != nil {
				t.Fatalf("marshal params: %v", err)
			}
		} else {
			paramsRaw = json.RawMessage("{}")
		}
		req, _ := json.Marshal(map[string]any{
			"jsonrpc": "2.0",
			"id":      msgID,
			"method":  method,
			"params":  paramsRaw,
		})
		if _, err := fmt.Fprintf(clientW, "Content-Length: %d\r\n\r\n", len(req)); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if _, err := clientW.Write(req); err != nil {
			t.Fatalf("write body: %v", err)
		}
		resp, err := readFrame(clientBuf)
		if err != nil {
			t.Fatalf("read response: %v", err)
		}
		return resp
	}

	closeFunc = func() {
		cancel()
		clientW.Close()
		serverR.Close()
		clientR.Close()
		serverW.Close()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("ServeConn did not stop within 2 seconds")
		}
	}

	return send, closeFunc
}

// readFrame reads one Content-Length-framed message from r.
func readFrame(r *bufio.Reader) (json.RawMessage, error) {
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
				return nil, fmt.Errorf("invalid Content-Length: %q", val)
			}
			contentLength = n
		}
	}
	if contentLength < 0 {
		return nil, errors.New("missing Content-Length")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(r, body); err != nil {
		return nil, err
	}
	return body, nil
}

func TestInitializeHandshake(t *testing.T) {
	srv := mcpstdio.NewServer("test-server", "0.1.0")
	send, close := testClient(t, srv)
	defer close()

	resp := send("initialize", map[string]any{"protocolVersion": "2024-11-05"})
	var result struct {
		Result struct {
			Capabilities struct {
				Tools any `json:"tools"`
			} `json:"capabilities"`
			ProtocolVersion string `json:"protocolVersion"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Result.Capabilities.Tools == nil {
		t.Error("expected capabilities.tools key to be present")
	}
	if result.Result.ProtocolVersion == "" {
		t.Error("expected protocolVersion in result")
	}
}

func TestToolsListEmpty(t *testing.T) {
	srv := mcpstdio.NewServer("test", "0.0.1")
	send, close := testClient(t, srv)
	defer close()

	resp := send("tools/list", nil)
	var result struct {
		Result struct {
			Tools []any `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Result.Tools == nil {
		t.Error("expected tools array (even if empty), got nil")
	}
}

func TestToolsListNonEmpty(t *testing.T) {
	srv := mcpstdio.NewServer("test", "0.0.1")
	srv.AddTool(mcpstdio.ToolDef{
		Name:        "echo",
		Description: "echo back",
		InputSchema: json.RawMessage(`{"type":"object","properties":{"text":{"type":"string"}}}`),
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		var p struct{ Text string }
		json.Unmarshal(args, &p) //nolint:errcheck
		return p.Text, nil
	})
	send, close := testClient(t, srv)
	defer close()

	resp := send("tools/list", nil)
	var result struct {
		Result struct {
			Tools []struct {
				Name        string          `json:"name"`
				InputSchema json.RawMessage `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(result.Result.Tools) == 0 {
		t.Fatal("expected non-empty tools list")
	}
	if result.Result.Tools[0].Name != "echo" {
		t.Errorf("tool name: got %q, want %q", result.Result.Tools[0].Name, "echo")
	}
	if len(result.Result.Tools[0].InputSchema) == 0 {
		t.Error("expected inputSchema to be set")
	}
}

func TestToolsCallSuccess(t *testing.T) {
	srv := mcpstdio.NewServer("test", "0.0.1")
	srv.AddTool(mcpstdio.ToolDef{
		Name:        "greet",
		Description: "say hello",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "hello world", nil
	})
	send, close := testClient(t, srv)
	defer close()

	resp := send("tools/call", map[string]any{
		"name":      "greet",
		"arguments": map[string]any{},
	})
	var result struct {
		Result struct {
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Result.IsError {
		t.Error("unexpected isError == true")
	}
	if len(result.Result.Content) == 0 {
		t.Fatal("expected content entries")
	}
	if result.Result.Content[0].Text != "hello world" {
		t.Errorf("content text: got %q, want %q", result.Result.Content[0].Text, "hello world")
	}
	if result.Result.Content[0].Type != "text" {
		t.Errorf("content type: got %q, want %q", result.Result.Content[0].Type, "text")
	}
}

func TestToolsCallHandlerError(t *testing.T) {
	srv := mcpstdio.NewServer("test", "0.0.1")
	srv.AddTool(mcpstdio.ToolDef{
		Name:        "fail",
		Description: "always fails",
		InputSchema: json.RawMessage(`{"type":"object"}`),
	}, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "", errors.New("something went wrong")
	})
	send, close := testClient(t, srv)
	defer close()

	resp := send("tools/call", map[string]any{
		"name":      "fail",
		"arguments": map[string]any{},
	})
	var result struct {
		Result struct {
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.Result.IsError {
		t.Error("expected isError == true")
	}
}

func TestToolsCallUnknownTool(t *testing.T) {
	srv := mcpstdio.NewServer("test", "0.0.1")
	send, close := testClient(t, srv)
	defer close()

	resp := send("tools/call", map[string]any{
		"name":      "does-not-exist",
		"arguments": map[string]any{},
	})
	var result struct {
		Error *struct {
			Code int `json:"code"`
		} `json:"error"`
		Result *struct {
			IsError bool `json:"isError"`
		} `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	// Unknown tool is a JSON-RPC protocol error, not an isError result.
	if result.Error == nil {
		t.Error("expected JSON-RPC error response for unknown tool, not isError")
	}
}

func TestUnknownMethod(t *testing.T) {
	srv := mcpstdio.NewServer("test", "0.0.1")
	send, close := testClient(t, srv)
	defer close()

	resp := send("no/such/method", nil)
	var result struct {
		Error *struct {
			Code int `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.Error == nil {
		t.Error("expected JSON-RPC error for unknown method")
	}
	if result.Error.Code != -32601 {
		t.Errorf("error code: got %d, want -32601", result.Error.Code)
	}
}

func TestNotificationsInitialized(t *testing.T) {
	srv := mcpstdio.NewServer("test", "0.0.1")

	serverR, clientW := io.Pipe()
	clientR, serverW := io.Pipe()

	ctx := t.Context()
	defer clientW.Close()
	defer serverR.Close()
	defer clientR.Close()
	defer serverW.Close()

	go srv.ServeConn(ctx, serverR, serverW) //nolint:errcheck

	writeRaw := func(v any) {
		body, _ := json.Marshal(v)
		fmt.Fprintf(clientW, "Content-Length: %d\r\n\r\n", len(body))
		clientW.Write(body) //nolint:errcheck
	}

	// Send notifications/initialized — no "id" field means it is a notification.
	writeRaw(map[string]any{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
		"params":  map[string]any{},
	})

	// Immediately send a real request.
	writeRaw(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]any{},
	})

	// We expect exactly one response (for tools/list, not for the notification).
	done := make(chan json.RawMessage, 1)
	go func() {
		resp, err := readFrame(bufio.NewReader(clientR))
		if err == nil {
			done <- resp
		}
	}()

	select {
	case resp := <-done:
		var result struct {
			Result *struct {
				Tools any `json:"tools"`
			} `json:"result"`
		}
		if err := json.Unmarshal(resp, &result); err != nil || result.Result == nil {
			t.Fatalf("expected tools/list response, got: %s", resp)
		}
	case <-time.After(2 * time.Second):
		t.Error("timed out: expected exactly one response (for tools/list)")
	}
}
