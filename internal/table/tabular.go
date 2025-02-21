// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

//go:build tabular || (!tablewriter && !termtables)
// +build tabular !tablewriter,!termtables

// See also version command's tabular.go and replicate build tag constraints there.

/*
Package table provides table support for character.
This implementation uses tabular for layout.

With the more modern Golang features, this should probably be in /internal/
namespace.
*/
package table

import (
	tb "go.pennock.tech/tabular/auto"
	"go.pennock.tech/tabular/properties"
	"go.pennock.tech/tabular/properties/align"
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
// columns starting from 1.
func (t *Table) AlignColumn(column int, ourAlign Alignment) {
	var how align.Alignment
	switch ourAlign {
	case LEFT:
		how = align.Left
	case CENTER:
		how = align.Center
	case RIGHT:
		how = align.Right
	}
	t.t.Column(column).SetProperty(align.PropertyType, how)
}

// SetSkipableColumn sets a column as skipable in some contexts (typically if
// every entry is empty).
func (t *Table) SetSkipableColumn(column int) {
	t.t.Column(column).SetProperty(properties.Skipable, true)
}

// We support styles.
func init() {
	AvailableStyles = tb.ListStyles()
	RenderStyle = defaultRenderStyle
}
