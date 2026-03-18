// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver

import (
	"testing"
	"time"

	"github.com/philpennock/character/sources"
)

func TestCursorRoundTrip(t *testing.T) {
	cases := []cursorData{
		{Type: "search", Query: "arrow", Offset: 0},
		{Type: "search", Query: "rightwards", Offset: 200},
		{Type: "name", Query: "check", Offset: 50},
		{Type: "block", Block: "Dingbats", Offset: 100},
	}
	for _, c := range cases {
		encoded := encodeCursor(c)
		decoded, err := decodeCursor(encoded)
		if err != nil {
			t.Errorf("decodeCursor(%q): %v", encoded, err)
			continue
		}
		if decoded.Type != c.Type || decoded.Query != c.Query || decoded.Block != c.Block || decoded.Offset != c.Offset {
			t.Errorf("round-trip mismatch: got %+v; want %+v", decoded, c)
		}
	}
}

func TestCursorCorruptedBase64(t *testing.T) {
	_, err := decodeCursor("!!!not-base64!!!")
	if err == nil {
		t.Error("expected error for corrupted base64")
	}
}

func TestCursorInvalidJSON(t *testing.T) {
	// Valid base64 but not valid JSON.
	_, err := decodeCursor("bm90LWpzb24")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCursorMissingType(t *testing.T) {
	// Encode a cursor with no Type field.
	_, err := decodeCursor("eyJvIjowfQ") // {"o":0}
	if err == nil {
		t.Error("expected error for missing type")
	}
}

func TestCacheBasic(t *testing.T) {
	c := newResultCache(1 * time.Minute)
	runes := []rune{'A', 'B', 'C'}
	c.put("test", runes)

	got, ok := c.get("test")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 3 || got[0] != 'A' {
		t.Errorf("got %v; want [A B C]", got)
	}
}

func TestCacheMiss(t *testing.T) {
	c := newResultCache(1 * time.Minute)
	_, ok := c.get("nonexistent")
	if ok {
		t.Error("expected cache miss")
	}
}

func TestCacheExpiry(t *testing.T) {
	c := newResultCache(1 * time.Millisecond)
	c.put("test", []rune{'A'})
	time.Sleep(5 * time.Millisecond)
	_, ok := c.get("test")
	if ok {
		t.Error("expected cache miss after TTL expiry")
	}
}

func testSrcs(t *testing.T) *sources.Sources {
	t.Helper()
	return sources.NewFast()
}

func TestPaginateRunesFull(t *testing.T) {
	srcs := testSrcs(t)
	runes := []rune{0x2713, 0x2714, 0x2715} // CHECK MARK, HEAVY CHECK MARK, MULTIPLICATION X

	result, err := paginateRunes(runes, pageParams{Detail: "full", Limit: 2}, cursorData{Type: "search", Query: "check"}, srcs)
	if err != nil {
		t.Fatal(err)
	}

	// Quick check: result should contain "count":2 and "total":3
	if !contains(result, `"count":2`) {
		t.Errorf("expected count:2 in %s", result)
	}
	if !contains(result, `"total":3`) {
		t.Errorf("expected total:3 in %s", result)
	}
	if !contains(result, `"cursor":`) {
		t.Errorf("expected cursor in %s", result)
	}
}

func TestPaginateRunesSummary(t *testing.T) {
	srcs := testSrcs(t)
	runes := []rune{0x2713, 0x2714}

	result, err := paginateRunes(runes, pageParams{Detail: "summary", Limit: 10}, cursorData{Type: "search", Query: "check"}, srcs)
	if err != nil {
		t.Fatal(err)
	}

	if !contains(result, `"columns":`) {
		t.Errorf("expected columns in summary result: %s", result)
	}
	if !contains(result, `"rows":`) {
		t.Errorf("expected rows in summary result: %s", result)
	}
	if contains(result, `"results":`) {
		t.Errorf("summary should not contain results: %s", result)
	}
}

func TestPaginateRunesNoCursor(t *testing.T) {
	srcs := testSrcs(t)
	runes := []rune{0x2713}

	result, err := paginateRunes(runes, pageParams{Detail: "full", Limit: 10}, cursorData{Type: "search", Query: "check"}, srcs)
	if err != nil {
		t.Fatal(err)
	}

	if contains(result, `"cursor":`) {
		t.Errorf("expected no cursor for single-page result: %s", result)
	}
}

func TestPaginateRunesCursorTypeMismatch(t *testing.T) {
	srcs := testSrcs(t)
	runes := []rune{0x2713}

	// Create a cursor for type "search" but pass cursorBase with type "block"
	cursor := encodeCursor(cursorData{Type: "search", Query: "check", Offset: 0})
	_, err := paginateRunes(runes, pageParams{Detail: "full", Limit: 10, Cursor: cursor}, cursorData{Type: "block", Block: "Dingbats"}, srcs)
	if err == nil {
		t.Error("expected error for cursor type mismatch")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
