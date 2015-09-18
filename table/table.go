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

func (t *Table) Empty() bool {
	return t.rowCount == 0
}

func (t *Table) Render() string {
	return t.t.Render()
}
