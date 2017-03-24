// Copyright © 2015-2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package resultset

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"

	"golang.org/x/net/idna"

	"github.com/philpennock/character/aux"
	"github.com/philpennock/character/entities"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/table"
	"github.com/philpennock/character/unicode"
)

// CanTable is the interface for the caller to determine if we have
// table-support loaded at all.  It mostly just avoids propagating imports of
// the table shim into every place which is already creating results.
func CanTable() bool {
	return table.Supported()
}

type selector int

// These constants dictate what is being added to a resultSet.
const (
	_ITEM selector = iota
	_ERROR
	_DIVIDER
)

type printItem uint

// These constants dictate what attribute of a rune should be printed.
const (
	PRINT_RUNE printItem = iota
	PRINT_RUNE_ISOLATED
	PRINT_RUNE_DEC
	PRINT_RUNE_HEX     // raw hex
	PRINT_RUNE_JSON    // surrogate pair in JSON syntax
	PRINT_RUNE_UTF8ENC // URL format
	PRINT_RUNE_PUNY
	PRINT_RUNE_WIDTH // best guess, terminal display cell width
	PRINT_NAME
	PRINT_BLOCK
	PRINT_HTML_ENTITIES
	PRINT_XML_ENTITIES
)

type fieldSetSelector uint

const (
	FIELD_SET_DEFAULT fieldSetSelector = iota
	FIELD_SET_NET
	FIELD_SET_DEBUG
)

type errorItem struct {
	input string
	err   error
}

type resultSet struct {
	sources *sources.Sources
	items   []unicode.CharInfo
	errors  []errorItem
	which   []selector

	OutputStream io.Writer
	ErrorStream  io.Writer

	// This is subject to change; do we want fully selectable sets of fields,
	// just pre-canned, something else?  For now ... let's keep it simple.
	fields fieldSetSelector
}

// New creates a resultSet, which records items and errors encountered, and a
// little structure, so that the results can be printed out in a variety of
// styles later.  Just the character, or tables of attributes, are derived from
// the recorded results.
func New(s *sources.Sources, sizeHint int) *resultSet {
	return &resultSet{
		sources: s,
		items:   make([]unicode.CharInfo, 0, sizeHint),
		errors:  make([]errorItem, 0, 3),
		which:   make([]selector, 0, sizeHint),
	}
}

// SelectFieldsNet says to use the network fields, not the default fields.
// This API call is very much subject to change.
func (rs *resultSet) SelectFieldsNet() {
	rs.fields = FIELD_SET_NET
}

// SelectFieldsDebug says to show some internal diagnostics, not the default fields.
// This API call is very much subject to change.
func (rs *resultSet) SelectFieldsDebug() {
	rs.fields = FIELD_SET_DEBUG
}

// AddError records, in-sequence, that we got an error at this point.
func (rs *resultSet) AddError(input string, e error) {
	rs.errors = append(rs.errors, errorItem{input, e})
	rs.which = append(rs.which, _ERROR)
}

// AddCharInfo is used for recording character information as an item in the result set.
func (rs *resultSet) AddCharInfo(ci unicode.CharInfo) {
	rs.items = append(rs.items, ci)
	rs.which = append(rs.which, _ITEM)
}

// AddDivider is use between words.
func (rs *resultSet) AddDivider() {
	rs.which = append(rs.which, _DIVIDER)
}

// ErrorCount sums the number of errors in the entire resultSet.
func (rs *resultSet) ErrorCount() int {
	return len(rs.errors)
}

func (rs *resultSet) fixStreams() {
	if rs.OutputStream == nil {
		rs.OutputStream = os.Stdout
	}
	if rs.ErrorStream == nil {
		rs.ErrorStream = os.Stderr
	}
}

// PrintPlain shows just characters, but with full errors interleaved too.
// One character or error per line.
func (rs *resultSet) PrintPlain(what printItem) {
	rs.fixStreams()
	var ii, ei int
	var s selector
	for _, s = range rs.which {
		switch s {
		case _ITEM:
			fmt.Fprintf(rs.OutputStream, "%s\n", rs.RenderCharInfoItem(rs.items[ii], what))
			ii++
		case _ERROR:
			fmt.Fprintf(rs.ErrorStream, "looking up %q: %s\n", rs.errors[ei].input, rs.errors[ei].err)
			ei++
		case _DIVIDER:
			fmt.Fprintln(rs.OutputStream)
		default:
			fmt.Fprintf(rs.ErrorStream, "internal error, unhandled item to print, of type %v", s)
		}
	}
}

// StringPlain returns the characters as chars in a word, dividers as a space.
func (rs *resultSet) StringPlain(what printItem) string {
	rs.fixStreams()
	out := make([]string, 0, len(rs.which))
	var ii, ei int
	var s selector
	for _, s = range rs.which {
		switch s {
		case _ITEM:
			out = append(out, rs.RenderCharInfoItem(rs.items[ii], what))
			ii++
		case _ERROR:
			fmt.Fprintf(rs.ErrorStream, "looking up %q: %s\n", rs.errors[ei].input, rs.errors[ei].err)
			ei++
		case _DIVIDER:
			out = append(out, " ")
		default:
			fmt.Fprintf(rs.ErrorStream, "internal error, unhandled item to print, of type %v", s)
		}
	}
	return strings.Join(out, "")
}

// LenTotalCount yields how many rows are in the resultset, including dividers and errors
func (rs *resultSet) LenTotalCount() int {
	return len(rs.which)
}

// LenItemCount yields how many successful items are in the table
func (rs *resultSet) LenItemCount() int {
	return len(rs.items)
}

// RenderCharInfoItem converts a unicode.CharInfo and an attribute selector into a string
func (rs *resultSet) RenderCharInfoItem(ci unicode.CharInfo, what printItem) string {
	// we use 0 as a special-case for things like combinations, where there's only a name
	if ci.Number == 0 && what != PRINT_NAME {
		return " "
	}
	switch what {
	case PRINT_RUNE:
		return string(ci.Number)
	case PRINT_RUNE_ISOLATED: // BROKEN
		// FIXME: None of these are actually working
		return fmt.Sprintf("%c%c%c",
			0x202A, // LEFT-TO-RIGHT EMBEDDING
			// 0x202D, // LEFT-TO-RIGHT OVERRIDE
			// 0x2066, // LEFT-TO-RIGHT ISOLATE
			ci.Number,
			// 0x2069, // POP DIRECTIONAL ISOLATE
			0x202C, // POP DIRECTIONAL FORMATTING
		)
	case PRINT_RUNE_HEX:
		return strconv.FormatUint(uint64(ci.Number), 16)
	case PRINT_RUNE_DEC:
		return strconv.FormatUint(uint64(ci.Number), 10)
	case PRINT_RUNE_UTF8ENC:
		bb := []byte(string(ci.Number))
		var s string
		for i := range bb {
			s += fmt.Sprintf("%%%X", bb[i])
		}
		return s
	case PRINT_RUNE_JSON:
		r1, r2 := utf16.EncodeRune(ci.Number)
		if r1 == 0xFFFD && r2 == 0xFFFD {
			if ci.Number <= 0xFFFF {
				return fmt.Sprintf("\\u%04X", ci.Number)
			}
			return "?"
		}
		return fmt.Sprintf("\\u%04X\\u%04X", r1, r2)
	case PRINT_RUNE_PUNY:
		p, err := idna.ToASCII(string(ci.Number))
		if err != nil {
			return ""
		}
		return p
	case PRINT_RUNE_WIDTH:
		return strconv.FormatUint(uint64(aux.DisplayCellWidth(string(ci.Number))), 10)
	case PRINT_NAME:
		return ci.Name
	case PRINT_BLOCK:
		return rs.sources.UBlocks.Lookup(ci.Number)
	case PRINT_HTML_ENTITIES:
		eList, ok := entities.HTMLEntitiesReverse[ci.Number]
		if !ok {
			return ""
		}
		return "&" + strings.Join(eList, "; &") + ";"
	case PRINT_XML_ENTITIES:
		eList, ok := entities.XMLEntitiesReverse[ci.Number]
		if !ok {
			return ""
		}
		return "&" + strings.Join(eList, "; &") + ";"
	default:
		panic(fmt.Sprintf("unhandled item to print: %v", what))
	}
}

// PrintTables provides much more verbose details about the contents of
// a resultSet, in a structured terminal table.
func (rs *resultSet) PrintTables() {
	rs.fixStreams()
	if len(rs.items) > 0 {
		t := table.New()
		t.AddHeaders(rs.detailsHeaders()...)
		ii := 0
		for _, s := range rs.which {
			switch s {
			case _ITEM:
				t.AddRow(rs.detailsFor(rs.items[ii])...)
				ii++
			case _ERROR:
				// skip, print in separate table below
			case _DIVIDER:
				t.AddSeparator()
			}
		}
		for _, align := range rs.detailsColumnAlignments() {
			t.AlignColumn(align.column, align.where)
		}
		fmt.Fprint(rs.OutputStream, t.Render())
	}
	if len(rs.errors) > 0 {
		t := table.New()
		t.AddHeaders("Problem Input", "Error")
		for _, ei := range rs.errors {
			t.AddRow(ei.input, ei.err)
		}
		fmt.Fprint(rs.ErrorStream, t.Render())
	}
}

func (rs *resultSet) detailsHeaders() []interface{} {
	switch rs.fields {
	case FIELD_SET_DEFAULT:
		return []interface{}{
			"C", "Name", "Hex", "Dec", "Block", "Vim", "HTML", "XML",
		}
	case FIELD_SET_NET:
		return []interface{}{
			"C", "Name", "Hex", "UTF-8", "JSON", "Punycode",
		}
	case FIELD_SET_DEBUG:
		return []interface{}{
			"C", "Width", "Hex", "Name",
		}
	}
	return nil
}

type columnAlignments struct {
	column int // 1-based
	where  table.Alignment
}

func (rs *resultSet) detailsColumnAlignments() []columnAlignments {
	switch rs.fields {
	case FIELD_SET_DEFAULT:
		return []columnAlignments{
			{3, table.RIGHT}, // Hex
			{4, table.RIGHT}, // Dec
			{5, table.RIGHT}, // Block??
		}
	case FIELD_SET_NET:
		return []columnAlignments{
			{3, table.RIGHT}, // Hex
			{4, table.RIGHT}, // Dec
		}
	case FIELD_SET_DEBUG:
		return []columnAlignments{
			{3, table.RIGHT}, // Hex
		}
	}
	return nil
}

func (rs *resultSet) detailsFor(ci unicode.CharInfo) []interface{} {
	switch rs.fields {
	case FIELD_SET_DEFAULT:
		return []interface{}{
			rs.RenderCharInfoItem(ci, PRINT_RUNE), // should be PRINT_RUNE_ISOLATED
			rs.RenderCharInfoItem(ci, PRINT_NAME),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_HEX),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_DEC),
			rs.RenderCharInfoItem(ci, PRINT_BLOCK),
			// We might put Info in here, to match old Perl script behaviour
			rs.sources.Vim.DigraphsFor(ci.Number),
			rs.RenderCharInfoItem(ci, PRINT_HTML_ENTITIES),
			rs.RenderCharInfoItem(ci, PRINT_XML_ENTITIES),
		}
	case FIELD_SET_NET:
		return []interface{}{
			rs.RenderCharInfoItem(ci, PRINT_RUNE), // should be PRINT_RUNE_ISOLATED
			rs.RenderCharInfoItem(ci, PRINT_NAME),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_HEX),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_UTF8ENC),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_JSON),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_PUNY),
		}
	case FIELD_SET_DEBUG:
		return []interface{}{
			rs.RenderCharInfoItem(ci, PRINT_RUNE), // should be PRINT_RUNE_ISOLATED
			rs.RenderCharInfoItem(ci, PRINT_RUNE_WIDTH),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_HEX),
			rs.RenderCharInfoItem(ci, PRINT_NAME),
		}
	}
	return nil
}
