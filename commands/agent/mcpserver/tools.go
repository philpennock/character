// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/philpennock/character/commands/transform"
	"github.com/philpennock/character/internal/mcpstdio"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"
)

// jsonResult marshals v to a JSON string for return from a tool handler.
func jsonResult(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}
	return string(b), nil
}

// registerTools registers all eight MCP tools on srv.
func registerTools(srv *mcpstdio.Server, srcs *sources.Sources, searchReady <-chan struct{}) {
	srv.AddTool(mcpstdio.ToolDef{
		Name:        "unicode_lookup_char",
		Description: "Look up a single Unicode character and return its full property object",
		InputSchema: schemaLookupChar,
	}, handleLookupChar(srcs))

	srv.AddTool(mcpstdio.ToolDef{
		Name:        "unicode_lookup_name",
		Description: "Look up a Unicode character by name (exact or substring search)",
		InputSchema: schemaLookupName,
	}, handleLookupName(srcs, searchReady))

	srv.AddTool(mcpstdio.ToolDef{
		Name:        "unicode_search",
		Description: "Search for Unicode characters whose names contain the query string",
		InputSchema: schemaSearch,
	}, handleSearch(srcs, searchReady))

	srv.AddTool(mcpstdio.ToolDef{
		Name:        "unicode_lookup_codepoint",
		Description: "Look up a Unicode character by codepoint (U+XXXX, 0xXXXX, or decimal)",
		InputSchema: schemaLookupCodepoint,
	}, handleLookupCodepoint(srcs))

	srv.AddTool(mcpstdio.ToolDef{
		Name:        "unicode_browse_block",
		Description: "Return all characters in the named Unicode block (max 3000)",
		InputSchema: schemaBrowseBlock,
	}, handleBrowseBlock(srcs))

	srv.AddTool(mcpstdio.ToolDef{
		Name:        "unicode_list_blocks",
		Description: "Return an ordered list of all Unicode blocks with their codepoint ranges",
		InputSchema: schemaListBlocks,
	}, handleListBlocks(srcs))

	srv.AddTool(mcpstdio.ToolDef{
		Name:        "unicode_emoji_flag",
		Description: "Return the regional-indicator pair for a two-letter country code flag emoji",
		InputSchema: schemaEmojiFlag,
	}, handleEmojiFlag(srcs))

	srv.AddTool(mcpstdio.ToolDef{
		Name:        "unicode_transform",
		Description: "Transform text using fraktur, math, scream, scream-decode, or turn",
		InputSchema: schemaTransform,
	}, handleTransform())
}

// --- tool handler closures ---

func handleLookupChar(srcs *sources.Sources) mcpstdio.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var p struct {
			Char string `json:"char"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
		if utf8.RuneCountInString(p.Char) != 1 {
			return "", fmt.Errorf("char must be exactly one Unicode codepoint, got %d in %q",
				utf8.RuneCountInString(p.Char), p.Char)
		}
		r, _ := utf8.DecodeRuneInString(p.Char)
		return jsonResult(CharPropsFromRune(r, srcs))
	}
}

func handleLookupName(srcs *sources.Sources, searchReady <-chan struct{}) mcpstdio.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var p struct {
			Name  string `json:"name"`
			Exact bool   `json:"exact"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		if p.Exact {
			upper := strings.ToUpper(p.Name)
			ci, ok := srcs.Unicode.ByName[upper]
			if !ok {
				return "", fmt.Errorf("no character named %q", p.Name)
			}
			return jsonResult([]CharProps{CharPropsFromRune(ci.Number, srcs)})
		}

		// Substring search: wait for the search index to be ready.
		if searchReady != nil {
			select {
			case <-searchReady:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		_, found := srcs.Unicode.Search.Query(p.Name, -1)
		if len(found) == 0 {
			return jsonResult([]CharProps{})
		}
		cis := charInfoSlice(found)
		return jsonResult(charPropsSlice(cis, srcs))
	}
}

func handleSearch(srcs *sources.Sources, searchReady <-chan struct{}) mcpstdio.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var p struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
		if searchReady != nil {
			select {
			case <-searchReady:
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
		_, found := srcs.Unicode.Search.Query(p.Query, -1)
		if len(found) == 0 {
			return jsonResult([]CharProps{})
		}
		return jsonResult(charPropsSlice(charInfoSlice(found), srcs))
	}
}

func handleLookupCodepoint(srcs *sources.Sources) mcpstdio.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var p struct {
			Codepoint string `json:"codepoint"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
		r, err := parseCodepoint(p.Codepoint)
		if err != nil {
			return "", err
		}
		return jsonResult(CharPropsFromRune(r, srcs))
	}
}

// parseCodepoint parses a codepoint string in U+XXXX, 0xXXXX, or decimal form.
func parseCodepoint(s string) (rune, error) {
	s = strings.TrimSpace(s)
	var n int64
	var err error
	switch {
	case strings.HasPrefix(s, "U+") || strings.HasPrefix(s, "u+"):
		n, err = strconv.ParseInt(s[2:], 16, 32)
	case strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X"):
		n, err = strconv.ParseInt(s[2:], 16, 32)
	default:
		n, err = strconv.ParseInt(s, 10, 32)
	}
	if err != nil {
		return 0, fmt.Errorf("cannot parse codepoint %q: %w", s, err)
	}
	return rune(n), nil
}

func handleBrowseBlock(srcs *sources.Sources) mcpstdio.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var p struct {
			Block string `json:"block"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		min, max, candidates := srcs.UBlocks.FindByName(p.Block)
		if min == 0 && max == 0 {
			if len(candidates) > 0 {
				return "", fmt.Errorf("ambiguous block name %q; candidates: %s",
					p.Block, strings.Join(candidates, ", "))
			}
			return "", fmt.Errorf("unknown block name %q; use unicode_list_blocks to see all block names", p.Block)
		}

		const limit = 3000
		var results []CharProps
		for r := min; r <= max; r++ {
			if _, ok := srcs.Unicode.ByRune[r]; !ok {
				continue
			}
			results = append(results, CharPropsFromRune(r, srcs))
			if len(results) > limit {
				return "", fmt.Errorf("block %q contains more than %d characters; use codepoint range directly", p.Block, limit)
			}
		}
		return jsonResult(results)
	}
}

func handleListBlocks(srcs *sources.Sources) mcpstdio.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		blocks := srcs.UBlocks.ListBlocks()
		result := make([]BlockObj, len(blocks))
		for i, bi := range blocks {
			result[i] = BlockObj{
				Name:  bi.Name,
				Start: fmt.Sprintf("U+%04X", bi.Min),
				End:   fmt.Sprintf("U+%04X", bi.Max),
			}
		}
		return jsonResult(result)
	}
}

// FlagResult is returned by unicode_emoji_flag.
type FlagResult struct {
	Indicator1 CharProps `json:"indicator_1"`
	Indicator2 CharProps `json:"indicator_2"`
	Combined   string    `json:"combined"`
}

func handleEmojiFlag(srcs *sources.Sources) mcpstdio.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var p struct {
			CountryCode string `json:"country_code"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
		cc := strings.ToUpper(strings.TrimSpace(p.CountryCode))
		if len(cc) != 2 {
			return "", fmt.Errorf("country_code must be exactly two letters, got %q", p.CountryCode)
		}
		for _, ch := range cc {
			if ch < 'A' || ch > 'Z' {
				return "", fmt.Errorf("country_code must contain only ASCII letters, got %q", p.CountryCode)
			}
		}
		r1 := rune(0x1F1E6 + (rune(cc[0]) - 'A'))
		r2 := rune(0x1F1E6 + (rune(cc[1]) - 'A'))
		return jsonResult(FlagResult{
			Indicator1: CharPropsFromRune(r1, srcs),
			Indicator2: CharPropsFromRune(r2, srcs),
			Combined:   string(r1) + string(r2),
		})
	}
}

// TransformResult is returned by unicode_transform.
type TransformResult struct {
	Input  string `json:"input"`
	Type   string `json:"type"`
	Output string `json:"output"`
}

func handleTransform() mcpstdio.Handler {
	return func(ctx context.Context, args json.RawMessage) (string, error) {
		var p struct {
			Type   string `json:"type"`
			Text   string `json:"text"`
			Target string `json:"target"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}

		var output string
		var err error
		switch p.Type {
		case "fraktur":
			output, err = transform.TransformFraktur([]string{p.Text})
		case "math":
			output, err = transform.TransformMath([]string{p.Text}, p.Target)
		case "scream":
			output = transform.NewEncoder().Replace(p.Text)
		case "scream-decode":
			output = transform.NewDecoder().Replace(p.Text)
		case "turn":
			output, err = transform.TransformTurn([]string{p.Text})
		default:
			return "", fmt.Errorf("unknown transform type %q; valid types: fraktur, math, scream, scream-decode, turn", p.Type)
		}
		if err != nil {
			return "", err
		}
		return jsonResult(TransformResult{Input: p.Text, Type: p.Type, Output: output})
	}
}

// charInfoSlice converts the []interface{} from Search.Query into []unicode.CharInfo.
func charInfoSlice(found []interface{}) unicode.CharInfoList {
	cis := make(unicode.CharInfoList, len(found))
	for i, item := range found {
		cis[i] = item.(unicode.CharInfo)
	}
	cis.Sort()
	return cis
}

// charPropsSlice converts a sorted list of CharInfo into []CharProps.
func charPropsSlice(cis unicode.CharInfoList, srcs *sources.Sources) []CharProps {
	result := make([]CharProps, len(cis))
	for i, ci := range cis {
		result[i] = CharPropsFromRune(ci.Number, srcs)
	}
	return result
}

