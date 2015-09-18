package table

import (
	"github.com/apcera/termtables"
)

type Table struct {
	t        *termtables.Table
	rowCount int
}

func New() *Table {
	t := termtables.CreateTable()
	t.UTF8Box()
	return &Table{t: t}
}

func (t *Table) AddHeaders(headers ...interface{}) {
	t.t.AddHeaders(headers...)
}

func (t *Table) AddRow(cells ...interface{}) {
	t.t.AddRow(cells...)
	t.rowCount += 1
}

func (t *Table) AddSeparator() {
	t.t.AddSeparator()
}

func (t *Table) Empty() bool {
	return t.rowCount == 0
}

func (t *Table) Render() string {
	return t.t.Render()
}

type Alignment int

const (
	LEFT Alignment = iota
	CENTER
	RIGHT
)

// AlignColumn counts columns starting from 1
func (t *Table) AlignColumn(column int, align Alignment) {
	// private type, can't declare variable, so call func via switch
	switch align {
	case LEFT:
		t.t.SetAlign(termtables.AlignLeft, column)
	case CENTER:
		t.t.SetAlign(termtables.AlignCenter, column)
	case RIGHT:
		t.t.SetAlign(termtables.AlignRight, column)
	default:
		panic("unhandled column alignment")
	}
}
