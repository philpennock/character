// Copyright © 2015-2017,2020-2022,2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package resultset

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf16"

	"golang.org/x/net/idna"

	"github.com/philpennock/character/entities"
	"github.com/philpennock/character/internal/runemanip"
	"github.com/philpennock/character/internal/table"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"
)

// CanTable is the interface for the caller to determine if we have
// table-support loaded at all.  It mostly just avoids propagating imports of
// the table shim into every place which is already creating results.
func CanTable() bool {
	return table.Supported()
}

type selector int

// These constants dictate what is being added to a ResultSet.
const (
	_ITEM selector = iota
	_STRSEQ
	_ERROR
	_DIVIDER
)

type printItem uint

// These constants dictate what attribute of a rune should be printed.
const (
	PRINT_RUNE printItem = iota
	PRINT_RUNE_ISOLATED
	PRINT_RUNE_PRESENT_TEXT
	PRINT_RUNE_PRESENT_EMOJI
	PRINT_RUNE_PRESENT_LEFT
	PRINT_RUNE_PRESENT_RIGHT
	PRINT_RUNE__RENDERERS // items before this render the rune itself
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
	PRINT_PART_OF // when we decomposed from input
)

type fieldSetSelector uint

// These constants determine which columns will appear in a table.
const (
	FIELD_SET_DEFAULT fieldSetSelector = iota
	FIELD_SET_NET
	FIELD_SET_DEBUG
)

// These are used to influence how runes are rendered
type runeRenderBiasType uint

const (
	runeRenderUnspecified runeRenderBiasType = iota
	runeRenderText
	runeRenderEmoji
	runeRenderLeft
	runeRenderRight
)

type errorItem struct {
	input string
	err   error
}

// Originally we stored unicode.CharInfo directly but as of the point where
// we support normalization handling, we also want to record where an item came
// from, because it might not be obvious from what was typed at the
// command-line.
// strseq is used only for _STRSEQ types
type charItem struct {
	unicode unicode.CharInfo
	partOf  rune
	strseq  string
}

// fixedWidthCell satisfies an interface used by the tabular table provider, to
// let us override the width. [types.go TerminalCellWidther -- not asserting here
// so as to maintain table provider abstraction]
//
// The table providers take various approaches to determining width; tabular uses
// go-runewidth which is Pretty Good and far better than most, but we are a tool
// all about Unicode and can sometimes do a little better in some corner cases.
type fixedWidthCell struct {
	s string
	w int
}

func (fwc fixedWidthCell) String() string         { return fwc.s }
func (fwc fixedWidthCell) TerminalCellWidth() int { return fwc.w }

type tcWidther interface {
	TerminalCellWidth() int
}

// A ResultSet is the collection of unicode characters (or near facsimiles thereof)
// about which we wish to see data.  Various front-end commands just figure out which
// characters are being asked about and throw them in the ResultSet, then at the end
// the ResultSet is asked to print itself in an appropriate format, which might be
// in a table, as lines, or in other ways.
// For convenience, errors are also accumulated here.  If emitting tables, the errors
// will be in a separate second table, but otherwise they're interleaved with
// the real characters in the correct order (but to stderr (probably), so if
// the streams diverge then this might not be reconstructible).
type ResultSet struct {
	sources *sources.Sources
	items   []charItem
	errors  []errorItem
	which   []selector

	OutputStream io.Writer
	ErrorStream  io.Writer

	// This is subject to change; do we want fully selectable sets of fields,
	// just pre-canned, something else?  For now ... let's keep it simple.
	fields fieldSetSelector

	// runeBias is used by the flags-control in this package, to select
	// UTS#51 presentation selectors to emit after a character, to try to bias how
	// it is shown.
	runeBias runeRenderBiasType
}

// New creates a ResultSet.
// We now make ResultSet an exported type, ugh, so this stutters when used.
// Most usage should never do that.
func New(s *sources.Sources, sizeHint int) *ResultSet {
	r := &ResultSet{
		sources: s,
		items:   make([]charItem, 0, sizeHint),
		errors:  make([]errorItem, 0, 3),
		which:   make([]selector, 0, sizeHint),
	}
	if ResultCmdFlags.Text {
		r.runeBias = runeRenderText
	}
	if ResultCmdFlags.Emoji {
		r.runeBias = runeRenderEmoji
	}
	if ResultCmdFlags.Left {
		r.runeBias = runeRenderLeft
	}
	if ResultCmdFlags.Right {
		r.runeBias = runeRenderRight
	}
	return r
}

// SelectFieldsNet says to use the network fields, not the default fields.
// This API call is very much subject to change.
func (rs *ResultSet) SelectFieldsNet() {
	rs.fields = FIELD_SET_NET
}

// SelectFieldsDebug says to show some internal diagnostics, not the default fields.
// This API call is very much subject to change.
func (rs *ResultSet) SelectFieldsDebug() {
	rs.fields = FIELD_SET_DEBUG
}

// RunePrintType returns PRINT_RUNE or one of its variants, appropriate to
// handle command-line-chosen variant selectors accordingly.
func (rs *ResultSet) RunePrintType() printItem {
	switch rs.runeBias {
	case runeRenderUnspecified:
		return PRINT_RUNE
	case runeRenderText:
		return PRINT_RUNE_PRESENT_TEXT
	case runeRenderEmoji:
		return PRINT_RUNE_PRESENT_EMOJI
	case runeRenderLeft:
		return PRINT_RUNE_PRESENT_LEFT
	case runeRenderRight:
		return PRINT_RUNE_PRESENT_RIGHT
	default:
		panic("unhandled internal runeBias")
	}
}

// AddError records, in-sequence, that we got an error at this point.
func (rs *ResultSet) AddError(input string, e error) {
	rs.errors = append(rs.errors, errorItem{input, e})
	rs.which = append(rs.which, _ERROR)
}

// AddCharInfo is used for recording character information as an item in the result set.
func (rs *ResultSet) AddCharInfo(ci unicode.CharInfo) {
	rs.items = append(rs.items, charItem{unicode: ci})
	rs.which = append(rs.which, _ITEM)
}

// AddStringSequence is used where we have some combining sequence to be rendered as one
func (rs *ResultSet) AddStringSequence(s string) {
	rs.items = append(rs.items, charItem{strseq: s})
	rs.which = append(rs.which, _STRSEQ)
}

// AddCharInfoDerivedFrom is used when the character was decomposed by us, so
// that we can display original input if requested.
func (rs *ResultSet) AddCharInfoDerivedFrom(ci unicode.CharInfo, from rune) {
	rs.items = append(rs.items, charItem{unicode: ci, partOf: from})
	rs.which = append(rs.which, _ITEM)
}

// AddDivider is use between words.
func (rs *ResultSet) AddDivider() {
	rs.which = append(rs.which, _DIVIDER)
}

// ErrorCount sums the number of errors in the entire ResultSet.
func (rs *ResultSet) ErrorCount() int {
	return len(rs.errors)
}

func (rs *ResultSet) fixStreams() {
	if rs.OutputStream == nil {
		rs.OutputStream = os.Stdout
	}
	if rs.ErrorStream == nil {
		rs.ErrorStream = os.Stderr
	}
}

// PrintPlain shows just characters, but with full errors interleaved too.
// One character or error per line.
func (rs *ResultSet) PrintPlain(what printItem) {
	rs.fixStreams()
	var ii, ei int
	var s selector
	for _, s = range rs.which {
		switch s {
		case _ITEM:
			fmt.Fprintf(rs.OutputStream, "%s\n", rs.RenderCharInfoItem(rs.items[ii], what))
			ii++
		case _STRSEQ:
			fmt.Fprintf(rs.OutputStream, "%s\n", rs.items[ii].strseq)
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
func (rs *ResultSet) StringPlain(what printItem) string {
	rs.fixStreams()
	out := make([]string, 0, len(rs.which))
	var ii, ei int
	var s selector
	for _, s = range rs.which {
		switch s {
		case _ITEM:
			item := rs.RenderCharInfoItem(rs.items[ii], what)
			if itemS, ok := item.(fmt.Stringer); ok {
				out = append(out, itemS.String())
			} else if itemS, ok := item.(string); ok {
				out = append(out, itemS)
			} else {
				// shouldn't happen (but can't have a primitive type satisfy an interface)
				out = append(out, fmt.Sprintf("%v", item))
			}
			ii++
		case _STRSEQ:
			// We don't double-print, the items in this should already have been emitted
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
func (rs *ResultSet) LenTotalCount() int {
	return len(rs.which)
}

// LenItemCount yields how many successful items are in the table
func (rs *ResultSet) LenItemCount() int {
	return len(rs.items)
}

// RenderCharInfoItem converts a charItem and an attribute selector into a string or Stringer
func (rs *ResultSet) RenderCharInfoItem(ci charItem, what printItem) any {
	var strForm string
	// Exceptional cases first:
	//
	// We use 0 as a special-case for things like combinations, where there's only a name:
	if ci.unicode.Number == 0 {
		if ci.strseq != "" {
			switch what {
			case PRINT_RUNE, PRINT_RUNE_PRESENT_TEXT, PRINT_RUNE_PRESENT_EMOJI, PRINT_RUNE_PRESENT_LEFT, PRINT_RUNE_PRESENT_RIGHT:
				// We need to fall through to put the correct wrappings around it
				strForm = ci.strseq
			case PRINT_NAME:
				return "(derived sequence)"
			default:
				return " "
			}
		} else if what != PRINT_NAME {
			return " "
		}
	} else {
		strForm = string(ci.unicode.Number)

		// Some other substitutions:
		if what < PRINT_RUNE__RENDERERS && !strconv.IsGraphic(ci.unicode.Number) {
			// We need controls to not be printed, such as 0x98 "START OF STRING"
			// The first 20 points are easy, there are replacements for them.
			// strconv.IsGraphic is false for codepoints in newer Unicode than Go
			// stdlib handles, so we can't just default to the replacement
			// character.
			// We probably need to encode "control" etc as bits in a bitfield in ci.unicode.
			// For now, let's use a heuristic: all the "real" control characters are in
			// "Basic Latin" or "Latin-1 Supplement" and Unicode won't be assigning new
			// code-points there which Go won't recognise as non-graphic.
			switch {
			case ci.unicode.Number < 0x20:
				return string(rune(0x2400 + ci.unicode.Number)) // 'Control Pictures' block
			case ci.unicode.Number <= 0xFF:
				return string(rune(0xFFFD)) // "REPLACEMENT CHARACTER"
			}
		}
	}

	// Deal with overrides for combining, for the various print forms
	var (
		width        int
		useFixedCell bool
	)
	switch what {
	case PRINT_RUNE, PRINT_RUNE_PRESENT_TEXT, PRINT_RUNE_PRESENT_EMOJI, PRINT_RUNE_PRESENT_LEFT, PRINT_RUNE_PRESENT_RIGHT:
		var override bool
		if width, override = runemanip.DisplayCellWidth(strForm); override {
			if width == 0 {
				// overrode to 0, probably is IsVariationSelector, don't fix width to that
				strForm = " " + strForm
				width = 1
			}
			useFixedCell = true
		}
		if ci.unicode.Number == 0 {
			// We have a string given in, it's combined and in my experience, "probably wrong" in cell width.
			// This is very heuristic and might create as many problems as it solves, so might be reverted.
			if width == 1 {
				width = 0
			}
		}
		if width == 0 {
			// Combining characters, don't combine with the table padding but with an explicit space.
			strForm = " " + strForm
			width = 1
		}
	}

	// Normal handling:
	switch what {
	case PRINT_RUNE:
		if useFixedCell {
			return fixedWidthCell{s: strForm, w: width}
		}
		return strForm
	case PRINT_RUNE_ISOLATED: // BROKEN
		// FIXME: None of these are actually working
		sOuter := fmt.Sprintf("%c%s%c",
			0x202A, // LEFT-TO-RIGHT EMBEDDING
			// 0x202D, // LEFT-TO-RIGHT OVERRIDE
			// 0x2066, // LEFT-TO-RIGHT ISOLATE
			strForm,
			// 0x2069, // POP DIRECTIONAL ISOLATE
			0x202C, // POP DIRECTIONAL FORMATTING
		)
		if useFixedCell {
			return fixedWidthCell{s: sOuter, w: width}
		}
		return sOuter

	// These next four demonstrate that we need a bit of a design rethink here, because
	// "2.11 Order of Emoji ZWJ Sequences" of
	// <http://unicode.org/reports/tr51/> demonstrates that we've got a cascade
	// of modifiers we could be applying:
	//
	// > When representing emoji ZWJ sequences for an individual person, the following order should be used:
	// > Order	Category										Section
	// > 1		Base											Section 1.4.1 Emoji Characters
	// > 2		Emoji modifier or emoji presentation selector	Section 2.4 Diversity
	// > 3		Hair component									Section 2.8 Hair Component
	// > 4		Color											Section 2.9, Color
	// > 5		Gender sign or object							Section 2.3.1, Gender-Neutral Emoji
	// > 6		Direction indicator								Section 2.10, Emoji Glyph Facing Direction
	//
	// Fortunately for me right now, the Emoji Glyph Facing Direction sequences
	// end with the emoji presentation selector so I don't need to rewrite for
	// the bits we support here.

	case PRINT_RUNE_PRESENT_TEXT: // text presentation selector, UTS#51 §1.4.3 ED-8
		if unicode.Emojiable(ci.unicode.Number) {
			strForm = strForm + "\uFE0E" // VARIATION SELECTOR-15
		}
		return fixedWidthCell{s: strForm, w: width}
	case PRINT_RUNE_PRESENT_EMOJI: // emoji presentation selector, UTS#51 §1.4.3 ED-9
		if unicode.Emojiable(ci.unicode.Number) {
			strForm = strForm + "\uFE0F" // VARIATION SELECTOR-16
		}
		return fixedWidthCell{s: strForm, w: width}
	case PRINT_RUNE_PRESENT_LEFT: // emoji glyph facing direction UTS#51 §2.10
		if unicode.Emojiable(ci.unicode.Number) {
			strForm = strForm + "\u200D\u2B05\uFE0F" // ZERO WIDTH JOINER, LEFTWARDS BLACK ARROW, VARIATION SELECTOR-16
		}
		return fixedWidthCell{s: strForm, w: width}
	case PRINT_RUNE_PRESENT_RIGHT: // emoji glyph facing direction UTS#51 §2.10
		if unicode.Emojiable(ci.unicode.Number) {
			strForm = strForm + "\u200D\u27A1\uFE0F" // ZERO WIDTH JOINER, BLACK RIGHTWARDS ARROW, VARIATION SELECTOR-16
		}
		return fixedWidthCell{s: strForm, w: width}

	case PRINT_RUNE_HEX:
		return strconv.FormatUint(uint64(ci.unicode.Number), 16)
	case PRINT_RUNE_DEC:
		return strconv.FormatUint(uint64(ci.unicode.Number), 10)
	case PRINT_RUNE_UTF8ENC:
		bb := []byte(string(ci.unicode.Number))
		var s string
		for i := range bb {
			s += fmt.Sprintf("%%%X", bb[i])
		}
		return s
	case PRINT_RUNE_JSON:
		r1, r2 := utf16.EncodeRune(ci.unicode.Number)
		if r1 == 0xFFFD && r2 == 0xFFFD {
			if ci.unicode.Number <= 0xFFFF {
				return fmt.Sprintf("\\u%04X", ci.unicode.Number)
			}
			return "?"
		}
		return fmt.Sprintf("\\u%04X\\u%04X", r1, r2)
	case PRINT_RUNE_PUNY:
		p, err := idna.ToASCII(string(ci.unicode.Number))
		if err != nil {
			return ""
		}
		return p
	case PRINT_RUNE_WIDTH:
		// If we supported color etc, then this would be a good opportunity to
		// use the override return bool to color red or something.
		width, _ := runemanip.DisplayCellWidth(string(ci.unicode.Number))
		return strconv.FormatUint(uint64(width), 10)
	case PRINT_NAME:
		if ci.unicode.NameWidth == 0 {
			return ci.unicode.Name
		}
		return fixedWidthCell{s: ci.unicode.Name, w: ci.unicode.NameWidth}
	case PRINT_BLOCK:
		return rs.sources.UBlocks.Lookup(ci.unicode.Number)
	case PRINT_HTML_ENTITIES:
		eList, ok := entities.HTMLEntitiesReverse[ci.unicode.Number]
		if !ok {
			return ""
		}
		return "&" + strings.Join(eList, "; &") + ";"
	case PRINT_XML_ENTITIES:
		eList, ok := entities.XMLEntitiesReverse[ci.unicode.Number]
		if !ok {
			return ""
		}
		return "&" + strings.Join(eList, "; &") + ";"
	case PRINT_PART_OF:
		if ci.partOf == 0 {
			return ""
		}
		return string(ci.partOf)
	default:
		panic(fmt.Sprintf("unhandled item to print: %v", what))
	}
}

// JItem is how a character is represented in JSON output.
type JItem struct {
	Display      string   `json:"display"`
	DisplayText  string   `json:"displayText"`
	DisplayEmoji string   `json:"displayEmoji"`
	DisplayLeft  string   `json:"displayLeft"`
	DisplayRight string   `json:"displayRight"`
	Name         string   `json:"name"`
	Hex          string   `json:"hex"`
	Dec          string   `json:"decimal"`
	Block        string   `json:"block"`
	VIMDigraphs  []string `json:"vimDigraphs,omitempty"`
	X11Digraphs  []string `json:"x11Digraphs,omitempty"`
	HTMLEntities []string `json:"htmlEntities,omitempty"`
	XMLEntities  []string `json:"xmlEntities,omitempty"`
	UTF8         string   `json:"utf8"`
	JSONEscape   string   `json:"jsonEscape"`
	RenderWidth  int      `json:"renderWidth"`
	Puny         string   `json:"puny"`
	PartOf       string   `json:"part-of,omitempty"`
}

// JInfo is something which might be shown instead of JItem.
type JInfo struct {
	Comment string `json:"comment"`
}

// JCombination is something which might be shown instead of JItem
type JCombination struct {
	Display string `json:"display"`
	Derived bool   `json:"derived"`
}

// S converts to a string, for JSON
func S(x any) string {
	switch s := x.(type) {
	case string:
		return s
	case fixedWidthCell:
		return s.s
	case fmt.Stringer:
		return s.String()
	default:
		return ""
	}
}

// JItemWidth converts to a width, for JSON
func JItemWidth(x any) int {
	switch s := x.(type) {
	case string:
		width, _ := runemanip.DisplayCellWidth(s)
		return width
	case tcWidther:
		return s.TerminalCellWidth()
	case fmt.Stringer:
		width, _ := runemanip.DisplayCellWidth(s.String())
		return width
	default:
		return 0
	}
}

// JSONEntry constructs a probably-JItem struct for JSON rendering of a character.
func (rs *ResultSet) JSONEntry(ci charItem) any {
	html, _ := entities.HTMLEntitiesReverse[ci.unicode.Number]
	xml, _ := entities.XMLEntitiesReverse[ci.unicode.Number]

	if ci.unicode.Number == 0 {
		if ci.unicode.Name != "" {
			return &JInfo{Comment: ci.unicode.Name}
		}
		if ci.strseq != "" {
			return &JCombination{Display: ci.strseq, Derived: true}
		}
	}

	return &JItem{
		Display:      S(rs.RenderCharInfoItem(ci, PRINT_RUNE)),
		DisplayText:  S(rs.RenderCharInfoItem(ci, PRINT_RUNE_PRESENT_TEXT)),
		DisplayEmoji: S(rs.RenderCharInfoItem(ci, PRINT_RUNE_PRESENT_EMOJI)),
		DisplayLeft:  S(rs.RenderCharInfoItem(ci, PRINT_RUNE_PRESENT_LEFT)),
		DisplayRight: S(rs.RenderCharInfoItem(ci, PRINT_RUNE_PRESENT_RIGHT)),
		Name:         S(rs.RenderCharInfoItem(ci, PRINT_NAME)),
		Hex:          S(rs.RenderCharInfoItem(ci, PRINT_RUNE_HEX)),
		Dec:          S(rs.RenderCharInfoItem(ci, PRINT_RUNE_DEC)),
		Block:        S(rs.RenderCharInfoItem(ci, PRINT_BLOCK)),
		VIMDigraphs:  rs.sources.Vim.DigraphsSliceFor(ci.unicode.Number),
		X11Digraphs:  rs.sources.X11.DigraphsSliceFor(ci.unicode.Number),
		HTMLEntities: html,
		XMLEntities:  xml,
		UTF8:         S(rs.RenderCharInfoItem(ci, PRINT_RUNE_UTF8ENC)),
		JSONEscape:   S(rs.RenderCharInfoItem(ci, PRINT_RUNE_JSON)),
		RenderWidth:  JItemWidth(rs.RenderCharInfoItem(ci, PRINT_RUNE)),
		Puny:         S(rs.RenderCharInfoItem(ci, PRINT_RUNE_PUNY)),
		PartOf:       S(rs.RenderCharInfoItem(ci, PRINT_PART_OF)),
	}
}

// PrintJSON shows everything we know about each result, in JSON.
// Rather than array of char/div/error, we have two top-level arrays
// of results and a divider is represented by a nil item.  Not sure
// how friendly that is for arbitrary input, but it helps with syncing I think
// (but errors inline might help more).
//
// Note that this is distinct from using a JSON table renderer for the normal
// tables and is instead emitting completely different output.
func (rs *ResultSet) PrintJSON() {
	rs.fixStreams()
	type JError struct {
		Input string `json:"input"`
		Error string `json:"error"`
	}
	var output struct {
		Characters []any    `json:"characters,omitempty"`
		Errors     []JError `json:"errors,omitempty"`
	}
	if len(rs.items) > 0 {
		output.Characters = make([]any, 0, len(rs.items))
		ii := 0
		for _, s := range rs.which {
			switch s {
			case _ITEM, _STRSEQ:
				output.Characters = append(output.Characters, rs.JSONEntry(rs.items[ii]))
				ii++
			case _ERROR:
				// skip, handled below
			case _DIVIDER:
				output.Characters = append(output.Characters, nil)
			}
		}
	}
	if len(rs.errors) > 0 {
		output.Errors = make([]JError, 0, len(rs.errors))
		for _, ei := range rs.errors {
			output.Errors = append(output.Errors, JError{ei.input, ei.err.Error()})
		}
	}

	b, err := json.MarshalIndent(&output, "", "  ")
	if err != nil {
		fmt.Fprintln(rs.ErrorStream, err)
		return
	}
	rs.OutputStream.Write(b)
	rs.OutputStream.Write([]byte{'\n'})
}

// PrintTables provides much more verbose details about the contents of
// a ResultSet, in a structured terminal table.
func (rs *ResultSet) PrintTables() {
	rs.fixStreams()
	if len(rs.items) > 0 {
		t := table.New()
		t.AddHeaders(rs.detailsHeaders()...)
		rs.SetOmitColumns(t, ResultCmdFlags.OmitColumns)
		ii := 0
		for _, s := range rs.which {
			switch s {
			case _ITEM, _STRSEQ:
				t.AddRow(rs.detailsFor(rs.items[ii])...)
				ii++
			case _ERROR:
				// skip, print in separate table below
			case _DIVIDER:
				t.AddSeparator()
			}
		}
		for _, props := range rs.detailsColumnProperties() {
			if props.align != table.UNSET {
				t.AlignColumn(props.column, props.align)
			}
			if props.skipable {
				t.SetSkipableColumn(props.column)
			}
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

func (rs *ResultSet) detailsHeaders() []any {
	switch rs.fields {
	case FIELD_SET_DEFAULT:
		return []any{
			"C", "Name", "Hex", "Dec", "Block", "Vim", "X11", "HTML", "XML",
		}
	case FIELD_SET_NET:
		return []any{
			"C", "Name", "Hex", "UTF-8", "JSON", "Punycode", "Of",
		}
	case FIELD_SET_DEBUG:
		return []any{
			"C", "Width", "Hex", "Name", "C-Type",
		}
	}
	return nil
}

type columnProperties struct {
	column   int // 1-based
	align    table.Alignment
	skipable bool
}

func (rs *ResultSet) detailsColumnProperties() []columnProperties {
	switch rs.fields {
	case FIELD_SET_DEFAULT:
		return []columnProperties{
			{3, table.RIGHT, false}, // Hex
			{4, table.RIGHT, false}, // Dec
			{6, table.UNSET, true},  // Vim
			{7, table.UNSET, true},  // X11
			{8, table.UNSET, true},  // HTML
			{9, table.UNSET, true},  // XML
		}
	case FIELD_SET_NET:
		return []columnProperties{
			{3, table.RIGHT, false}, // Hex
			{4, table.RIGHT, false}, // Dec
			{7, table.UNSET, true},  // Of
		}
	case FIELD_SET_DEBUG:
		return []columnProperties{
			{3, table.RIGHT, false}, // Hex
		}
	}
	return nil
}

func (rs *ResultSet) detailsFor(ci charItem) []any {
	runeDisplay := PRINT_RUNE // should be PRINT_RUNE_ISOLATED
	switch rs.runeBias {
	case runeRenderUnspecified:
		// no action, want PRINT_RUNE
	case runeRenderText:
		runeDisplay = PRINT_RUNE_PRESENT_TEXT
	case runeRenderEmoji:
		runeDisplay = PRINT_RUNE_PRESENT_EMOJI
	case runeRenderLeft:
		runeDisplay = PRINT_RUNE_PRESENT_LEFT
	case runeRenderRight:
		runeDisplay = PRINT_RUNE_PRESENT_RIGHT

	}
	switch rs.fields {
	case FIELD_SET_DEFAULT:
		return []any{
			rs.RenderCharInfoItem(ci, runeDisplay),
			rs.RenderCharInfoItem(ci, PRINT_NAME),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_HEX),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_DEC),
			rs.RenderCharInfoItem(ci, PRINT_BLOCK),
			// We might put Info in here, to match old Perl script behaviour
			rs.sources.Vim.DigraphsFor(ci.unicode.Number),
			rs.sources.X11.DigraphsFor(ci.unicode.Number),
			rs.RenderCharInfoItem(ci, PRINT_HTML_ENTITIES),
			rs.RenderCharInfoItem(ci, PRINT_XML_ENTITIES),
			// PRINT_PART_OF is almost always empty and while important, it
			// annoys me, so I've removed it from my most-common view.
		}
	case FIELD_SET_NET:
		return []any{
			rs.RenderCharInfoItem(ci, runeDisplay),
			rs.RenderCharInfoItem(ci, PRINT_NAME),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_HEX),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_UTF8ENC),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_JSON),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_PUNY),
			rs.RenderCharInfoItem(ci, PRINT_PART_OF),
		}
	case FIELD_SET_DEBUG:
		return []any{
			rs.RenderCharInfoItem(ci, runeDisplay),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_WIDTH),
			rs.RenderCharInfoItem(ci, PRINT_RUNE_HEX),
			rs.RenderCharInfoItem(ci, PRINT_NAME),
			func(c charItem) string { t := rs.RenderCharInfoItem(ci, runeDisplay); return fmt.Sprintf("%T", t) }(ci),
		}
	}
	return nil
}

func (rs *ResultSet) SetOmitColumns(tab *table.Table, nameList string) {
	if nameList == "" {
		return
	}
	for _, field := range strings.Split(nameList, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}
		err := tab.SetOmitColumnName(field)
		if err != nil {
			rs.AddError("field "+strconv.Quote(field), err)
		}
	}
}
