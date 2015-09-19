package resultset

import (
	"fmt"
	"os"
	"strconv"

	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/table"
	"github.com/philpennock/character/unicode"
)

type selector int

const (
	ITEM selector = iota
	ERROR
	DIVIDER
)

type printItem int

const (
	PRINT_RUNE printItem = iota
	PRINT_RUNE_DEC
	PRINT_RUNE_HEX
	PRINT_RUNE_UTF8ENC
	PRINT_NAME
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
}

func New(s *sources.Sources, sizeHint int) *resultSet {
	return &resultSet{
		sources: s,
		items:   make([]unicode.CharInfo, 0, sizeHint),
		errors:  make([]errorItem, 0, 3),
		which:   make([]selector, 0, sizeHint),
	}
}

func (rs *resultSet) AddError(input string, e error) {
	rs.errors = append(rs.errors, errorItem{input, e})
	rs.which = append(rs.which, ERROR)
}

func (rs *resultSet) AddCharInfo(ci unicode.CharInfo) {
	rs.items = append(rs.items, ci)
	rs.which = append(rs.which, ITEM)
}

func (rs *resultSet) AddDivider() {
	rs.which = append(rs.which, DIVIDER)
}

func (rs *resultSet) ErrorCount() int {
	return len(rs.errors)
}

func (rs *resultSet) PrintPlain(what printItem) {
	var ii, ei int
	var s selector
	for _, s = range rs.which {
		switch s {
		case ITEM:
			fmt.Printf("%s\n", renderCharInfoItem(rs.items[ii], what))
			ii += 1
		case ERROR:
			fmt.Fprintf(os.Stderr, "looking up %q: %s\n", rs.errors[ei].input, rs.errors[ei].err)
			ei += 1
		case DIVIDER:
			fmt.Println()
		default:
			fmt.Fprintf(os.Stderr, "internal error, unhandled item to print, of type %v", s)
		}
	}
}

func renderCharInfoItem(ci unicode.CharInfo, what printItem) string {
	switch what {
	case PRINT_RUNE:
		return string(ci.Number)
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
	case PRINT_NAME:
		return ci.Name
	default:
		panic(fmt.Sprintf("unhandled item to print: %v", what))
	}
}

func (rs *resultSet) PrintTables() {
	if len(rs.items) > 0 {
		t := table.New()
		t.AddHeaders(detailsHeaders()...)
		ii := 0
		for _, s := range rs.which {
			switch s {
			case ITEM:
				t.AddRow(rs.detailsFor(rs.items[ii])...)
				ii += 1
			case ERROR:
				// skip, print in separate table below
			case DIVIDER:
				t.AddSeparator()
			}
		}
		for _, align := range detailsColumnAlignments {
			t.AlignColumn(align.column, align.where)
		}
		fmt.Print(t.Render())
	}
	if len(rs.errors) > 0 {
		t := table.New()
		t.AddHeaders("Problem Input", "Error")
		for _, ei := range rs.errors {
			t.AddRow(ei.input, ei.err)
		}
		fmt.Fprint(os.Stderr, t.Render())
	}
}

func detailsHeaders() []interface{} {
	return []interface{}{
		"C", "Name", "Hex", "Dec", "UTF-8", "Block", "Info", "Vim", "HTML", "XHTML",
	}
}

var detailsColumnAlignments = []struct {
	column int // 1-based
	where  table.Alignment
}{
	{3, table.RIGHT},
	{4, table.RIGHT},
	{5, table.RIGHT},
}

func (rs *resultSet) detailsFor(ci unicode.CharInfo) []interface{} {
	return []interface{}{
		renderCharInfoItem(ci, PRINT_RUNE),
		renderCharInfoItem(ci, PRINT_NAME),
		renderCharInfoItem(ci, PRINT_RUNE_HEX),
		renderCharInfoItem(ci, PRINT_RUNE_DEC),
		renderCharInfoItem(ci, PRINT_RUNE_UTF8ENC),
		// FIXME:
		"b?", "i?",
		rs.sources.Vim.DigraphsFor(ci.Number),
		"h?", "x?",
	}
}
