// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// +build tabular !tablewriter,!termtables

// See also version command's tabular.go and replicate build tag constraints there.

/*
This implementation uses tabular for layout.

Known current limitations:

* No columnar alignment
*/
package table

import (
	tb "go.pennock.tech/tabular/auto"
)

// Supported indicates that we have a terminal table provider loaded.
func Supported() bool { return true }

// defaultRenderStyle indicates which style to use by default.  Is used
// to initialize the RenderStyle control variable, which should be used
// for lookups.
const defaultRenderStyle = "utf8-heavy"

// A Table encapsulates our terminal table object from the dependency.
type Table struct {
	t        tb.RenderTable
	rowCount int
}

// New gives us a new empty table, configured for our basic requirements.
func New() *Table {
	return &Table{
		t: tb.New(RenderStyle),
	}
}

// AddHeaders takes a sequence of header-names for each column, and configures
// them as the header row.
func (t *Table) AddHeaders(headers ...interface{}) {
	t.t.AddHeaders(headers...)
}

// AddRow takes a sequence of cells for one table body row.
func (t *Table) AddRow(cells ...interface{}) {
	t.t.AddRowItems(cells...)
	t.rowCount++
}

// AddSeparator adds one row to the table, as a separator.
func (t *Table) AddSeparator() {
	t.t.AddSeparator()
}

// Empty is a predicate indicating if any real content has been added.
func (t *Table) Empty() bool {
	return t.rowCount == 0
}

// Render gives us the full table as a string for display.
// FIXME: why can't we return an error here?
func (t *Table) Render() string {
	rendered, err := t.t.Render()
	if err != nil {
		rendered += "\n" + err.Error() + "\n"
	}
	return rendered
}

// AlignColumn sets the alignment of one column in a given table.  It counts
// columns starting from 1.  In this implementation, it has no effect because
// tabular doesn't yet support alignment.
func (t *Table) AlignColumn(column int, align Alignment) {
	// IGNORED
	_ = column
	_ = align
}

// We support styles.
func init() {
	AvailableStyles = tb.ListStyles()
	RenderStyle = defaultRenderStyle
}
