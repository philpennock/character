// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package uformat_test

import (
	"testing"

	"github.com/philpennock/character/internal/uformat"
)

func TestUTF8Bytes(t *testing.T) {
	tests := []struct {
		r    rune
		want string
	}{
		{0x2713, "e2 9c 93"},     // CHECK MARK
		{0x1F600, "f0 9f 98 80"}, // GRINNING FACE
		{0x0041, "41"},           // LATIN CAPITAL A
		{0x007F, "7f"},           // DEL
		{0x0080, "c2 80"},        // first two-byte
	}
	for _, tt := range tests {
		got := uformat.UTF8Bytes(tt.r)
		if got != tt.want {
			t.Errorf("UTF8Bytes(U+%04X) = %q; want %q", tt.r, got, tt.want)
		}
	}
}

func TestUTF8Escaped(t *testing.T) {
	tests := []struct {
		r    rune
		want string
	}{
		{0x2713, `\xe2\x9c\x93`},
		{0x1F600, `\xf0\x9f\x98\x80`},
		{0x0041, `\x41`},
	}
	for _, tt := range tests {
		got := uformat.UTF8Escaped(tt.r)
		if got != tt.want {
			t.Errorf("UTF8Escaped(U+%04X) = %q; want %q", tt.r, got, tt.want)
		}
	}
}

func TestUnicodeEscaped(t *testing.T) {
	tests := []struct {
		r    rune
		want string
	}{
		{0x2713, `\u2713`},
		{0x1F600, `\U0001F600`},
		{0x0041, `\u0041`},
		{0xFFFF, `\uFFFF`},
		{0x10000, `\U00010000`},
	}
	for _, tt := range tests {
		got := uformat.UnicodeEscaped(tt.r)
		if got != tt.want {
			t.Errorf("UnicodeEscaped(U+%04X) = %q; want %q", tt.r, got, tt.want)
		}
	}
}

func TestRustEscaped(t *testing.T) {
	tests := []struct {
		r    rune
		want string
	}{
		{0x2713, `\u{2713}`},
		{0x1F600, `\u{1F600}`},
		{0x0041, `\u{41}`},
	}
	for _, tt := range tests {
		got := uformat.RustEscaped(tt.r)
		if got != tt.want {
			t.Errorf("RustEscaped(U+%04X) = %q; want %q", tt.r, got, tt.want)
		}
	}
}

func TestJSONEscaped(t *testing.T) {
	tests := []struct {
		r    rune
		want string
	}{
		{0x2713, `\u2713`},        // CHECK MARK — BMP
		{0x0041, `\u0041`},        // LATIN CAPITAL A
		{0x1F1FA, `\uD83C\uDDFA`}, // REGIONAL INDICATOR U — surrogate pair
		{0x1F600, `\uD83D\uDE00`}, // GRINNING FACE
	}
	for _, tt := range tests {
		got := uformat.JSONEscaped(tt.r)
		if got != tt.want {
			t.Errorf("JSONEscaped(U+%04X) = %q; want %q", tt.r, got, tt.want)
		}
	}
}
