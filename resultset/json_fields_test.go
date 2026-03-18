// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package resultset_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/philpennock/character/resultset"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"
)

// jsonOutputForRune runs the JSON output pipeline for a single rune and
// returns the first JItem in the characters array.
func jsonOutputForRune(t *testing.T, srcs *sources.Sources, r rune) *resultset.JItem {
	t.Helper()
	ci := unicode.CharInfo{Number: r}
	// Look up the real CharInfo from sources so Name etc. are populated.
	if info, ok := srcs.Unicode.ByRune[r]; ok {
		ci = info
	}

	rs := resultset.New(srcs, 1)
	rs.AddCharInfo(ci)

	var sb strings.Builder
	rs.OutputStream = &sb

	rs.PrintJSON()

	var output struct {
		Characters []json.RawMessage `json:"characters"`
	}
	if err := json.Unmarshal([]byte(sb.String()), &output); err != nil {
		t.Fatalf("unmarshal JSON output: %v\nraw: %s", err, sb.String())
	}
	if len(output.Characters) == 0 {
		t.Fatal("no characters in JSON output")
	}

	var item resultset.JItem
	if err := json.Unmarshal(output.Characters[0], &item); err != nil {
		t.Fatalf("unmarshal JItem: %v\nraw: %s", err, output.Characters[0])
	}
	return &item
}

func TestJItemNewFields(t *testing.T) {
	srcs := sources.NewFast()

	t.Run("check_mark_U2713", func(t *testing.T) {
		item := jsonOutputForRune(t, srcs, 0x2713)

		if item.UTF8Bytes != "e2 9c 93" {
			t.Errorf("UTF8Bytes = %q; want %q", item.UTF8Bytes, "e2 9c 93")
		}
		if item.UTF8Escaped != `\xe2\x9c\x93` {
			t.Errorf("UTF8Escaped = %q; want %q", item.UTF8Escaped, `\xe2\x9c\x93`)
		}
		if item.UnicodeEscaped != `\u2713` {
			t.Errorf("UnicodeEscaped = %q; want %q", item.UnicodeEscaped, `\u2713`)
		}
		if item.RustEscaped != `\u{2713}` {
			t.Errorf("RustEscaped = %q; want %q", item.RustEscaped, `\u{2713}`)
		}
		if item.Category != "So" {
			t.Errorf("Category = %q; want %q", item.Category, "So")
		}
		if item.BlockInfo == nil {
			t.Fatal("BlockInfo is nil")
		}
		if item.BlockInfo.Name != "Dingbats" {
			t.Errorf("BlockInfo.Name = %q; want %q", item.BlockInfo.Name, "Dingbats")
		}
		if item.BlockInfo.Start != "U+2700" {
			t.Errorf("BlockInfo.Start = %q; want %q", item.BlockInfo.Start, "U+2700")
		}
	})

	t.Run("grinning_face_U1F600", func(t *testing.T) {
		item := jsonOutputForRune(t, srcs, 0x1F600)

		if item.UnicodeEscaped != `\U0001F600` {
			t.Errorf("UnicodeEscaped = %q; want %q", item.UnicodeEscaped, `\U0001F600`)
		}
		if item.RustEscaped != `\u{1F600}` {
			t.Errorf("RustEscaped = %q; want %q", item.RustEscaped, `\u{1F600}`)
		}
		if item.UTF8Bytes != "f0 9f 98 80" {
			t.Errorf("UTF8Bytes = %q; want %q", item.UTF8Bytes, "f0 9f 98 80")
		}
	})

	t.Run("snowman_has_emoji_variant", func(t *testing.T) {
		item := jsonOutputForRune(t, srcs, 0x2603) // SNOWMAN
		if len(item.PresentVariants) == 0 {
			t.Fatal("PresentVariants is empty for SNOWMAN; expected at least emoji variant")
		}
		var found bool
		for _, pv := range item.PresentVariants {
			if pv.Selector == "U+FE0F" && pv.Type == "emoji" {
				found = true
			}
		}
		if !found {
			t.Errorf("PresentVariants = %v; want entry with Selector=U+FE0F, Type=emoji", item.PresentVariants)
		}
	})

	t.Run("latin_a_no_variants", func(t *testing.T) {
		item := jsonOutputForRune(t, srcs, 0x0041) // LATIN CAPITAL LETTER A
		if len(item.PresentVariants) != 0 {
			t.Errorf("PresentVariants = %v; want nil/empty for U+0041", item.PresentVariants)
		}
	})
}
