// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode_test

import (
	"testing"

	"github.com/philpennock/character/unicode"
)

func TestPresentationVariants(t *testing.T) {
	t.Run("snowman_has_emoji_variant", func(t *testing.T) {
		// U+2603 SNOWMAN has both text (FE0E) and emoji (FE0F) variants.
		variants := unicode.PresentationVariants(0x2603)
		if len(variants) == 0 {
			t.Fatal("PresentationVariants(U+2603 SNOWMAN) returned nil; expected at least emoji variant")
		}
		var found bool
		for _, v := range variants {
			if v.Selector == 0xFE0F && v.Type == "emoji" {
				found = true
			}
		}
		if !found {
			t.Errorf("PresentationVariants(U+2603) = %v; want entry with Selector=U+FE0F, Type=emoji", variants)
		}
	})

	t.Run("latin_a_has_no_variants", func(t *testing.T) {
		// U+0041 LATIN CAPITAL LETTER A has no variation selectors.
		variants := unicode.PresentationVariants(0x0041)
		if variants != nil {
			t.Errorf("PresentationVariants(U+0041) = %v; want nil", variants)
		}
	})

	t.Run("number_sign_has_both_variants", func(t *testing.T) {
		// U+0023 NUMBER SIGN (#) has both text and emoji variants in the data.
		variants := unicode.PresentationVariants(0x0023)
		if len(variants) < 2 {
			t.Fatalf("PresentationVariants(U+0023) = %v; want at least 2 entries (text+emoji)", variants)
		}
		var hasText, hasEmoji bool
		for _, v := range variants {
			switch v.Type {
			case "text":
				hasText = true
			case "emoji":
				hasEmoji = true
			}
		}
		if !hasText {
			t.Error("expected a 'text' variant for U+0023")
		}
		if !hasEmoji {
			t.Error("expected an 'emoji' variant for U+0023")
		}
	})
}
