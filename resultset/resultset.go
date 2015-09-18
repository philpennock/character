package resultset

import (
	"fmt"
	"os"
	"strconv"

	"github.com/philpennock/character/table"
	"github.com/philpennock/character/unicode"
)

type selector int

const (
	ITEM selector = iota
	ERROR
)

type printItem int

const (
	PRINT_RUNE printItem = iota
	PRINT_RUNE_DEC
	PRINT_RUNE_HEX
	PRINT_NAME
)

type errorItem struct {
	input string
	err   error
}

type resultSet struct {
	items  []unicode.CharInfo
	errors []errorItem
	which  []selector
}

func New(sizeHint int) *resultSet {
	return &resultSet{
		items:  make([]unicode.CharInfo, 0, sizeHint),
		errors: make([]errorItem, 0, sizeHint),
		which:  make([]selector, 0, sizeHint),
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
		for _, ci := range rs.items {
			t.AddRow(detailsFor(ci)...)
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
		"C", "Name", "Hex", "Dec", "Block", "Info", "Vim", "HTML", "XHTML",
	}
}

func detailsFor(ci unicode.CharInfo) []interface{} {
	return []interface{}{
		renderCharInfoItem(ci, PRINT_RUNE),
		renderCharInfoItem(ci, PRINT_NAME),
		renderCharInfoItem(ci, PRINT_RUNE_HEX),
		renderCharInfoItem(ci, PRINT_RUNE_DEC),
		// FIXME:
		"b?", "i?", "v?", "h?", "x?",
	}
}
