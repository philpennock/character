// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode_test

import (
	"testing"

	"github.com/philpennock/character/unicode"
)

func TestBlocksLookupInfo(t *testing.T) {
	blocks := unicode.LoadBlocks()

	t.Run("check_mark_in_dingbats", func(t *testing.T) {
		info := blocks.LookupInfo(0x2713) // CHECK MARK
		if info == nil {
			t.Fatal("LookupInfo(U+2713) returned nil")
		}
		if info.Name != "Dingbats" {
			t.Errorf("Name = %q; want %q", info.Name, "Dingbats")
		}
		if info.Min > 0x2713 || info.Max < 0x2713 {
			t.Errorf("block range [U+%04X..U+%04X] does not contain U+2713", info.Min, info.Max)
		}
	})

	t.Run("latin_capital_a", func(t *testing.T) {
		info := blocks.LookupInfo(0x0041) // LATIN CAPITAL LETTER A
		if info == nil {
			t.Fatal("LookupInfo(U+0041) returned nil")
		}
		if info.Name == "" {
			t.Error("block name should not be empty for U+0041")
		}
	})

	t.Run("high_codepoint_no_panic", func(t *testing.T) {
		// Should return nil (or a valid block), must not panic.
		_ = blocks.LookupInfo(0xFFFF)
	})

	t.Run("copy_independence", func(t *testing.T) {
		// Two calls should return independent copies.
		a := blocks.LookupInfo(0x2713)
		b := blocks.LookupInfo(0x2713)
		if a == b {
			t.Error("LookupInfo returned the same pointer twice; expected independent copies")
		}
	})
}
