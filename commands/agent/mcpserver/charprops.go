// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// Package mcpserver implements the MCP tool handlers for the character Unicode
// lookup server.  It does not import resultset or read ResultCmdFlags; all
// data comes from the sources package and uformat helpers.
package mcpserver

import (
	"fmt"
	"unicode/utf8"

	"github.com/philpennock/character/entities"
	"github.com/philpennock/character/internal/runemanip"
	"github.com/philpennock/character/internal/uformat"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"
)

// BlockObj holds structured Unicode block information in MCP output.
type BlockObj struct {
	Name  string `json:"name"`
	Start string `json:"start"` // "U+2700"
	End   string `json:"end"`   // "U+27BF"
}

// PresentVar describes one presentation-selector variant in MCP output.
type PresentVar struct {
	Selector string `json:"selector"` // "U+FE0F"
	Type     string `json:"type"`     // "text" or "emoji"
}

// CharProps is the full character property object returned by MCP lookup tools.
type CharProps struct {
	Character      string       `json:"character"`
	Name           string       `json:"name"`
	Hex            string       `json:"hex"`
	Decimal        int          `json:"decimal"`
	UTF8Percent    string       `json:"utf8_percent"`
	UTF8Bytes      string       `json:"utf8_bytes"`
	UTF8Escaped    string       `json:"utf8_escaped"`
	UnicodeEscaped string       `json:"unicode_escaped"`
	RustEscaped    string       `json:"rust_escaped"`
	JSONEscaped    string       `json:"json_escaped"`
	Block          BlockObj     `json:"block"`
	Category       string       `json:"category"`
	RenderWidth    int          `json:"render_width"`
	HTMLEntities   []string     `json:"html_entities,omitempty"`
	XMLEntities    []string     `json:"xml_entities,omitempty"`
	VimDigraphs    []string     `json:"vim_digraphs,omitempty"`
	X11Digraphs    []string     `json:"x11_digraphs,omitempty"`
	PresentVariants []PresentVar `json:"presentation_variants,omitempty"`
}

// utf8Percent returns URL-percent encoding for r (e.g. "%E2%9C%93").
func utf8Percent(r rune) string {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	s := ""
	for i := range n {
		s += fmt.Sprintf("%%%X", buf[i])
	}
	return s
}

// CharPropsFromRune computes all character properties for r from the given
// sources.  It must not read or write resultset.ResultCmdFlags.
func CharPropsFromRune(r rune, srcs *sources.Sources) CharProps {
	var name string
	if ci, ok := srcs.Unicode.ByRune[r]; ok {
		name = ci.Name
	}

	var block BlockObj
	if bi := srcs.UBlocks.LookupInfo(r); bi != nil {
		block = BlockObj{
			Name:  bi.Name,
			Start: fmt.Sprintf("U+%04X", bi.Min),
			End:   fmt.Sprintf("U+%04X", bi.Max),
		}
	}

	width, _ := runemanip.DisplayCellWidth(string(r))

	html, _ := entities.HTMLEntitiesReverse[r]
	xml, _ := entities.XMLEntitiesReverse[r]

	var presentVars []PresentVar
	for _, pv := range unicode.PresentationVariants(r) {
		presentVars = append(presentVars, PresentVar{
			Selector: fmt.Sprintf("U+%04X", pv.Selector),
			Type:     pv.Type,
		})
	}

	return CharProps{
		Character:      string(r),
		Name:           name,
		Hex:            fmt.Sprintf("%X", r),
		Decimal:        int(r),
		UTF8Percent:    utf8Percent(r),
		UTF8Bytes:      uformat.UTF8Bytes(r),
		UTF8Escaped:    uformat.UTF8Escaped(r),
		UnicodeEscaped: uformat.UnicodeEscaped(r),
		RustEscaped:    uformat.RustEscaped(r),
		JSONEscaped:    uformat.JSONEscaped(r),
		Block:          block,
		Category:       unicode.GeneralCategory(r),
		RenderWidth:    width,
		HTMLEntities:   html,
		XMLEntities:    xml,
		VimDigraphs:    srcs.Vim.DigraphsSliceFor(r),
		X11Digraphs:    srcs.X11.DigraphsSliceFor(r),
		PresentVariants: presentVars,
	}
}
