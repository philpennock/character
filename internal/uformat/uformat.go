// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// Package uformat provides pure rune→string formatting helpers for Unicode
// byte representations.  All functions are stateless and have no side effects.
package uformat

import (
	"unicode/utf16"
	"unicode/utf8"
)

const upperHex = "0123456789ABCDEF"
const lowerHex = "0123456789abcdef"

// hexWidth returns the number of hex digits needed to represent v,
// with a minimum of minWidth.
func hexWidth(v rune, minWidth int) int {
	w := 0
	for x := uint32(v); x > 0; x >>= 4 {
		w++
	}
	if w < minWidth {
		w = minWidth
	}
	return w
}

// putHex writes n uppercase hex digits of v into dst starting at off.
func putHex(dst []byte, off int, hex string, v uint32, n int) {
	for i := n - 1; i >= 0; i-- {
		dst[off+i] = hex[v&0xF]
		v >>= 4
	}
}

// Codepoint returns the Unicode codepoint notation for r, e.g. "U+0041" or
// "U+1F600".  Uses a minimum of four uppercase hex digits.
func Codepoint(r rune) string {
	w := hexWidth(r, 4)
	var buf [8]byte // "U+" (2) + up to 6 hex digits
	buf[0] = 'U'
	buf[1] = '+'
	putHex(buf[:], 2, upperHex, uint32(r), w)
	return string(buf[:2+w])
}

// HexUpper returns the minimum-width uppercase hex representation of r.
// For example, "41" for 'A', "1F600" for '😀'.
func HexUpper(r rune) string {
	w := hexWidth(r, 1)
	var buf [6]byte
	putHex(buf[:], 0, upperHex, uint32(r), w)
	return string(buf[:w])
}

// UTF8PercentEncoded returns the UTF-8 encoding of r as concatenated
// percent-encoded bytes, uppercase.  For example, U+2713 → "%E2%9C%93".
// This is the encoding used for non-ASCII bytes in URL percent-encoding.
func UTF8PercentEncoded(r rune) string {
	var enc [utf8.UTFMax]byte
	n := utf8.EncodeRune(enc[:], r)
	var buf [utf8.UTFMax * 3]byte // max 12
	for i := range n {
		buf[i*3+0] = '%'
		buf[i*3+1] = upperHex[enc[i]>>4]
		buf[i*3+2] = upperHex[enc[i]&0xF]
	}
	return string(buf[:n*3])
}

// UTF8Bytes returns the UTF-8 encoding of r as a space-separated lowercase
// two-digit hex byte string.  For example, U+2713 → "e2 9c 93".
func UTF8Bytes(r rune) string {
	var enc [utf8.UTFMax]byte
	n := utf8.EncodeRune(enc[:], r)
	var buf [utf8.UTFMax*3 - 1]byte // max "xx xx xx xx" = 11
	pos := 0
	for i := range n {
		if i > 0 {
			buf[pos] = ' '
			pos++
		}
		buf[pos] = lowerHex[enc[i]>>4]
		buf[pos+1] = lowerHex[enc[i]&0xF]
		pos += 2
	}
	return string(buf[:pos])
}

// UTF8Escaped returns the UTF-8 encoding of r as concatenated \xNN escape
// sequences, lowercase.  For example, U+2713 → "\\xe2\\x9c\\x93".
func UTF8Escaped(r rune) string {
	var enc [utf8.UTFMax]byte
	n := utf8.EncodeRune(enc[:], r)
	var buf [utf8.UTFMax * 4]byte // max 16
	for i := range n {
		buf[i*4+0] = '\\'
		buf[i*4+1] = 'x'
		buf[i*4+2] = lowerHex[enc[i]>>4]
		buf[i*4+3] = lowerHex[enc[i]&0xF]
	}
	return string(buf[:n*4])
}

// writeU4 writes \uXXXX (four uppercase hex digits) into dst at off, returning
// the position after the last byte written.
func writeU4(dst []byte, off int, v uint32) int {
	dst[off] = '\\'
	dst[off+1] = 'u'
	putHex(dst, off+2, upperHex, v, 4)
	return off + 6
}

// UnicodeEscaped returns the Unicode escape sequence for r.
// For r ≤ U+FFFF, returns \uXXXX (four uppercase hex digits).
// For r > U+FFFF, returns \UXXXXXXXX (eight uppercase hex digits).
func UnicodeEscaped(r rune) string {
	if r <= 0xFFFF {
		var buf [6]byte // \uXXXX
		writeU4(buf[:], 0, uint32(r))
		return string(buf[:])
	}
	var buf [10]byte // \UXXXXXXXX
	buf[0] = '\\'
	buf[1] = 'U'
	putHex(buf[:], 2, upperHex, uint32(r), 8)
	return string(buf[:])
}

// RustEscaped returns the Rust Unicode escape sequence for r, using the
// minimum-length uppercase hex representation without leading zeros.
// For example, U+2713 → "\\u{2713}", U+1F600 → "\\u{1F600}".
func RustEscaped(r rune) string {
	w := hexWidth(r, 1)
	var buf [10]byte // \u{10FFFF} = 10 max
	buf[0] = '\\'
	buf[1] = 'u'
	buf[2] = '{'
	putHex(buf[:], 3, upperHex, uint32(r), w)
	buf[3+w] = '}'
	return string(buf[:4+w])
}

// JSONEscaped returns the JSON/JavaScript escape sequence for r.
// For characters outside the BMP (r > U+FFFF), returns a UTF-16 surrogate
// pair \uXXXX\uXXXX.  For BMP characters, returns \uXXXX.
// This mirrors the PRINT_RUNE_JSON logic in resultset.
func JSONEscaped(r rune) string {
	r1, r2 := utf16.EncodeRune(r)
	if r1 == 0xFFFD && r2 == 0xFFFD {
		if r <= 0xFFFF {
			var buf [6]byte
			writeU4(buf[:], 0, uint32(r))
			return string(buf[:])
		}
		return "?"
	}
	var buf [12]byte // \uXXXX\uXXXX
	writeU4(buf[:], 0, uint32(r1))
	writeU4(buf[:], 6, uint32(r2))
	return string(buf[:])
}
