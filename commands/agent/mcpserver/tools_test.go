// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/philpennock/character/commands/agent/mcpserver"
	"github.com/philpennock/character/internal/mcpstdio"
	"github.com/philpennock/character/sources"
)

// toolClient provides a thin JSON-RPC client over a mcpstdio.Server for testing.
type toolClient struct {
	t      *testing.T
	write  *io.PipeWriter
	buf    *bufio.Reader
	cancel context.CancelFunc
	done   chan error
	msgID  int
}

func newToolClient(t *testing.T, inner *mcpstdio.Server) *toolClient {
	t.Helper()
	serverR, clientW := io.Pipe()
	clientR, serverW := io.Pipe()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- inner.ServeConn(ctx, serverR, serverW)
	}()

	tc := &toolClient{
		t:      t,
		write:  clientW,
		buf:    bufio.NewReader(clientR),
		cancel: cancel,
		done:   done,
	}

	// Perform initialize handshake.
	resp := tc.send("initialize", map[string]any{"protocolVersion": "2024-11-05"})
	var init struct {
		Result any `json:"result"`
	}
	if err := json.Unmarshal(resp, &init); err != nil || init.Result == nil {
		t.Fatalf("initialize handshake failed: %v / %s", err, resp)
	}

	t.Cleanup(func() {
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
	})

	return tc
}

func (tc *toolClient) send(method string, params any) json.RawMessage {
	tc.t.Helper()
	tc.msgID++
	paramsRaw, _ := json.Marshal(params)
	req, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      tc.msgID,
		"method":  method,
		"params":  json.RawMessage(paramsRaw),
	})
	fmt.Fprintf(tc.write, "%s\n", req)
	resp, err := readClientFrame(tc.buf)
	if err != nil {
		tc.t.Fatalf("read response: %v", err)
	}
	return resp
}

// callTool sends a tools/call request and returns the inner result object.
func (tc *toolClient) callTool(toolName string, args any) json.RawMessage {
	tc.t.Helper()
	resp := tc.send("tools/call", map[string]any{
		"name":      toolName,
		"arguments": args,
	})
	var result struct {
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(resp, &result); err != nil {
		tc.t.Fatalf("unmarshal response: %v\nraw: %s", err, resp)
	}
	return result.Result
}

func readClientFrame(r *bufio.Reader) (json.RawMessage, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	return line, nil
}

// callViaServeConn is a one-shot helper that creates a new client per call.
func callViaServeConn(t *testing.T, inner *mcpstdio.Server, toolName string, args any) (json.RawMessage, bool) {
	t.Helper()
	tc := newToolClient(t, inner)
	result := tc.callTool(toolName, args)
	return result, true
}

func newTestSrcs(t *testing.T) *sources.Sources {
	t.Helper()
	srcs := sources.NewFast()
	srcs.LoadUnicodeSearch()
	return srcs
}

// --- test helpers ---

func assertIsError(t *testing.T, result json.RawMessage, wantSubstr string) {
	t.Helper()
	var r struct {
		IsError bool `json:"isError"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !r.IsError {
		t.Errorf("expected isError=true, got false; content: %v", r.Content)
		return
	}
	if wantSubstr != "" && len(r.Content) > 0 {
		if !strings.Contains(strings.ToLower(r.Content[0].Text), strings.ToLower(wantSubstr)) {
			t.Errorf("error text %q does not contain %q", r.Content[0].Text, wantSubstr)
		}
	}
}

func extractCharProps(t *testing.T, result json.RawMessage) mcpserver.CharProps {
	t.Helper()
	var r struct {
		Content []struct{ Text string `json:"text"` } `json:"content"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(r.Content) == 0 {
		t.Fatal("no content")
	}
	var cp mcpserver.CharProps
	if err := json.Unmarshal([]byte(r.Content[0].Text), &cp); err != nil {
		t.Fatalf("unmarshal CharProps: %v\ntext: %s", err, r.Content[0].Text)
	}
	return cp
}

func extractCharPropsSlice(t *testing.T, result json.RawMessage) []mcpserver.CharProps {
	t.Helper()
	var r struct {
		Content []struct{ Text string `json:"text"` } `json:"content"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(r.Content) == 0 {
		t.Fatal("no content")
	}
	var cps []mcpserver.CharProps
	if err := json.Unmarshal([]byte(r.Content[0].Text), &cps); err != nil {
		t.Fatalf("unmarshal []CharProps: %v\ntext: %s", err, r.Content[0].Text)
	}
	return cps
}

func extractBlockSlice(t *testing.T, result json.RawMessage) []mcpserver.BlockObj {
	t.Helper()
	var r struct {
		Content []struct{ Text string `json:"text"` } `json:"content"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(r.Content) == 0 {
		t.Fatal("no content")
	}
	var blocks []mcpserver.BlockObj
	if err := json.Unmarshal([]byte(r.Content[0].Text), &blocks); err != nil {
		t.Fatalf("unmarshal []BlockObj: %v\ntext: %s", err, r.Content[0].Text)
	}
	return blocks
}

func extractFlagResult(t *testing.T, result json.RawMessage) mcpserver.FlagResult {
	t.Helper()
	var r struct {
		Content []struct{ Text string `json:"text"` } `json:"content"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(r.Content) == 0 {
		t.Fatal("no content")
	}
	var fr mcpserver.FlagResult
	if err := json.Unmarshal([]byte(r.Content[0].Text), &fr); err != nil {
		t.Fatalf("unmarshal FlagResult: %v\ntext: %s", err, r.Content[0].Text)
	}
	return fr
}

func extractTransformResult(t *testing.T, result json.RawMessage) mcpserver.TransformResult {
	t.Helper()
	var r struct {
		Content []struct{ Text string `json:"text"` } `json:"content"`
	}
	if err := json.Unmarshal(result, &r); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if len(r.Content) == 0 {
		t.Fatal("no content")
	}
	var tr mcpserver.TransformResult
	if err := json.Unmarshal([]byte(r.Content[0].Text), &tr); err != nil {
		t.Fatalf("unmarshal TransformResult: %v\ntext: %s", err, r.Content[0].Text)
	}
	return tr
}

// --- tool tests ---

func TestToolLookupChar(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_lookup_char", map[string]any{"char": "✓"})
	cp := extractCharProps(t, result)
	if cp.Name != "CHECK MARK" {
		t.Errorf("Name = %q; want %q", cp.Name, "CHECK MARK")
	}
}

func TestToolLookupCharEmpty(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_lookup_char", map[string]any{"char": ""})
	assertIsError(t, result, "exactly one")
}

func TestToolLookupCharTwoRunes(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_lookup_char", map[string]any{"char": "ab"})
	assertIsError(t, result, "")
}

func TestToolLookupCodepointUPlus(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_lookup_codepoint", map[string]any{"codepoint": "U+2713"})
	cp := extractCharProps(t, result)
	if cp.Name != "CHECK MARK" {
		t.Errorf("Name = %q; want CHECK MARK", cp.Name)
	}
}

func TestToolLookupCodepointDecimal(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_lookup_codepoint", map[string]any{"codepoint": "10003"})
	cp := extractCharProps(t, result)
	if cp.Name != "CHECK MARK" {
		t.Errorf("Name = %q; want CHECK MARK", cp.Name)
	}
}

func TestToolLookupCodepointHex(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_lookup_codepoint", map[string]any{"codepoint": "0x2713"})
	cp := extractCharProps(t, result)
	if cp.Name != "CHECK MARK" {
		t.Errorf("Name = %q; want CHECK MARK", cp.Name)
	}
}

func TestToolLookupNameExact(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_lookup_name",
		map[string]any{"name": "CHECK MARK", "exact": true})
	cps := extractCharPropsSlice(t, result)
	if len(cps) != 1 || cps[0].Name != "CHECK MARK" {
		t.Errorf("got %v; want single CHECK MARK", cps)
	}
}

func TestToolLookupNameExactNotFound(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_lookup_name",
		map[string]any{"name": "NONEXISTENT XYZ CHARACTER", "exact": true})
	assertIsError(t, result, "")
}

func TestToolSearch(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_search", map[string]any{"query": "snowman"})
	cps := extractCharPropsSlice(t, result)
	if len(cps) == 0 {
		t.Fatal("expected search results for 'snowman'")
	}
	var found bool
	for _, cp := range cps {
		if cp.Name == "SNOWMAN" {
			found = true
		}
	}
	if !found {
		t.Errorf("SNOWMAN not in search results")
	}
}

func TestToolListBlocks(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_list_blocks", map[string]any{})
	blocks := extractBlockSlice(t, result)
	if len(blocks) == 0 {
		t.Fatal("expected non-empty blocks list")
	}
	for i, b := range blocks {
		if b.Name == "" {
			t.Errorf("blocks[%d].Name is empty", i)
		}
		if len(b.Start) < 3 || b.Start[0] != 'U' || b.Start[1] != '+' {
			t.Errorf("blocks[%d].Start = %q; want U+XXXX format", i, b.Start)
		}
	}
}

func TestToolBrowseBlock(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_browse_block", map[string]any{"block": "Dingbats"})
	cps := extractCharPropsSlice(t, result)
	if len(cps) == 0 {
		t.Fatal("expected non-empty Dingbats block")
	}
}

func TestToolBrowseBlockNotFound(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_browse_block",
		map[string]any{"block": "Nonexistent Block XYZ"})
	assertIsError(t, result, "unknown block")
}

func TestToolEmojiFlagFR(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_emoji_flag", map[string]any{"country_code": "FR"})
	fr := extractFlagResult(t, result)
	if fr.Combined == "" {
		t.Error("Combined is empty")
	}
}

func TestToolEmojiFlagLowercase(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()

	upperResult, _ := callViaServeConn(t, inner, "unicode_emoji_flag", map[string]any{"country_code": "FR"})
	lowerResult, _ := callViaServeConn(t, inner, "unicode_emoji_flag", map[string]any{"country_code": "fr"})

	fu := extractFlagResult(t, upperResult)
	fl := extractFlagResult(t, lowerResult)
	if fu.Combined != fl.Combined {
		t.Errorf("FR=%q != fr=%q; expected same flag", fu.Combined, fl.Combined)
	}
}

func TestToolEmojiFlagThreeLetters(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_emoji_flag", map[string]any{"country_code": "ZZZ"})
	assertIsError(t, result, "exactly two letters")
}

func TestToolTransformFraktur(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_transform",
		map[string]any{"type": "fraktur", "text": "Hello"})
	tr := extractTransformResult(t, result)
	if tr.Output == tr.Input || tr.Output == "" {
		t.Errorf("fraktur output %q should differ from input %q", tr.Output, tr.Input)
	}
}

func TestToolTransformScreamRoundtrip(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	orig := "hello world"

	encResult, _ := callViaServeConn(t, inner, "unicode_transform",
		map[string]any{"type": "scream", "text": orig})
	enc := extractTransformResult(t, encResult)

	decResult, _ := callViaServeConn(t, inner, "unicode_transform",
		map[string]any{"type": "scream-decode", "text": enc.Output})
	dec := extractTransformResult(t, decResult)

	if dec.Output != orig {
		t.Errorf("scream roundtrip: got %q; want %q", dec.Output, orig)
	}
}

func TestToolTransformInvalidType(t *testing.T) {
	srcs := newTestSrcs(t)
	inner := mcpserver.NewServer(srcs, nil).Inner()
	result, _ := callViaServeConn(t, inner, "unicode_transform",
		map[string]any{"type": "invalid", "text": "hello"})
	assertIsError(t, result, "unknown transform type")
}
