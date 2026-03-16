// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// Package uformat provides pure rune→string formatting helpers for Unicode
// byte representations.  All functions are stateless and have no side effects.
package uformat

import (
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// UTF8Bytes returns the UTF-8 encoding of r as a space-separated lowercase
// two-digit hex byte string.  For example, U+2713 → "e2 9c 93".
func UTF8Bytes(r rune) string {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	parts := make([]string, n)
	for i := range n {
		parts[i] = fmt.Sprintf("%02x", buf[i])
	}
	return strings.Join(parts, " ")
}

// UTF8Escaped returns the UTF-8 encoding of r as concatenated \xNN escape
// sequences, lowercase.  For example, U+2713 → "\\xe2\\x9c\\x93".
func UTF8Escaped(r rune) string {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	var sb strings.Builder
	sb.Grow(n * 4)
	for i := range n {
		fmt.Fprintf(&sb, "\\x%02x", buf[i])
	}
	return sb.String()
}

// UnicodeEscaped returns the Unicode escape sequence for r.
// For r ≤ U+FFFF, returns \uXXXX (four uppercase hex digits).
// For r > U+FFFF, returns \UXXXXXXXX (eight uppercase hex digits).
func UnicodeEscaped(r rune) string {
	if r <= 0xFFFF {
		return fmt.Sprintf("\\u%04X", r)
	}
	return fmt.Sprintf("\\U%08X", r)
}

// RustEscaped returns the Rust Unicode escape sequence for r, using the
// minimum-length uppercase hex representation without leading zeros.
// For example, U+2713 → "\\u{2713}", U+1F600 → "\\u{1F600}".
func RustEscaped(r rune) string {
	return fmt.Sprintf("\\u{%X}", r)
}

// JSONEscaped returns the JSON/JavaScript escape sequence for r.
// For characters outside the BMP (r > U+FFFF), returns a UTF-16 surrogate
// pair \uXXXX\uXXXX.  For BMP characters, returns \uXXXX.
// This mirrors the PRINT_RUNE_JSON logic in resultset.
func JSONEscaped(r rune) string {
	r1, r2 := utf16.EncodeRune(r)
	if r1 == 0xFFFD && r2 == 0xFFFD {
		if r <= 0xFFFF {
			return fmt.Sprintf("\\u%04X", r)
		}
		return "?"
	}
	return fmt.Sprintf("\\u%04X\\u%04X", r1, r2)
}
