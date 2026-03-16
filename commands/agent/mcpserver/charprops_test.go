// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver_test

import (
	"testing"

	"github.com/philpennock/character/commands/agent/mcpserver"
	"github.com/philpennock/character/sources"
)

func TestCharPropsFromRune(t *testing.T) {
	srcs := sources.NewFast()

	t.Run("check_mark_U2713", func(t *testing.T) {
		cp := mcpserver.CharPropsFromRune(0x2713, srcs)
		if cp.Name != "CHECK MARK" {
			t.Errorf("Name = %q; want %q", cp.Name, "CHECK MARK")
		}
		if cp.Hex != "2713" {
			t.Errorf("Hex = %q; want %q", cp.Hex, "2713")
		}
		if cp.Decimal != 10003 {
			t.Errorf("Decimal = %d; want %d", cp.Decimal, 10003)
		}
		if cp.UTF8Bytes != "e2 9c 93" {
			t.Errorf("UTF8Bytes = %q; want %q", cp.UTF8Bytes, "e2 9c 93")
		}
		if cp.UnicodeEscaped != `\u2713` {
			t.Errorf("UnicodeEscaped = %q; want %q", cp.UnicodeEscaped, `\u2713`)
		}
		if cp.RustEscaped != `\u{2713}` {
			t.Errorf("RustEscaped = %q; want %q", cp.RustEscaped, `\u{2713}`)
		}
		if cp.Block.Name != "Dingbats" {
			t.Errorf("Block.Name = %q; want %q", cp.Block.Name, "Dingbats")
		}
		if cp.Block.Start != "U+2700" {
			t.Errorf("Block.Start = %q; want %q", cp.Block.Start, "U+2700")
		}
		if cp.Category != "So" {
			t.Errorf("Category = %q; want %q", cp.Category, "So")
		}
		if cp.RenderWidth != 1 {
			t.Errorf("RenderWidth = %d; want 1", cp.RenderWidth)
		}
	})

	t.Run("latin_capital_a_U0041", func(t *testing.T) {
		cp := mcpserver.CharPropsFromRune(0x0041, srcs)
		if cp.Name != "LATIN CAPITAL LETTER A" {
			t.Errorf("Name = %q; want %q", cp.Name, "LATIN CAPITAL LETTER A")
		}
		if cp.Category != "Lu" {
			t.Errorf("Category = %q; want %q", cp.Category, "Lu")
		}
	})

	t.Run("grinning_face_U1F600", func(t *testing.T) {
		cp := mcpserver.CharPropsFromRune(0x1F600, srcs)
		if cp.UnicodeEscaped != `\U0001F600` {
			t.Errorf("UnicodeEscaped = %q; want %q", cp.UnicodeEscaped, `\U0001F600`)
		}
		if cp.RustEscaped != `\u{1F600}` {
			t.Errorf("RustEscaped = %q; want %q", cp.RustEscaped, `\u{1F600}`)
		}
		if cp.UTF8Bytes != "f0 9f 98 80" {
			t.Errorf("UTF8Bytes = %q; want %q", cp.UTF8Bytes, "f0 9f 98 80")
		}
	})

	t.Run("regional_indicator_block", func(t *testing.T) {
		cp := mcpserver.CharPropsFromRune(0x1F1EB, srcs) // REGIONAL INDICATOR LETTER F
		if !containsCI(cp.Block.Name, "enclosed alphanumeric") {
			t.Errorf("Block.Name = %q; expected to contain 'enclosed alphanumeric' (case-insensitive)", cp.Block.Name)
		}
	})
}

func containsCI(s, substr string) bool {
	return len(s) >= len(substr) && containsLower(lower(s), lower(substr))
}

func lower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		b[i] = c
	}
	return string(b)
}

func containsLower(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
