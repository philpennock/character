// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

/*
Package table is the internal shim for providing table support for laying out
content prettily.

All interaction with the dependency which provides terminal-drawn tables
should go through this module.  This provides a shim, giving us isolation.
We know the exact subset of features which we rely upon, and can switch
providers.  If desired, we can use multiple files with build-tags, to let the
dependency be satisfied at build-time.
*/
package table

import (
	"github.com/apcera/termtables"
)

// Supported indicates that we have a terminal table provider loaded.
// This exists in anticipation of multiple providers and tables being optional
// if not supported at all.
func Supported() bool { return true }

// A Table encapsulates our terminal table object from the dependency.
type Table struct {
	t        *termtables.Table
	rowCount int
}

// New gives us a new empty table, configured for our basic requirements.
func New() *Table {
	t := termtables.CreateTable()
	t.UTF8Box()
	return &Table{t: t}
}

// AddHeaders takes a sequence of header-names for each column, and configures
// them as the header row.
func (t *Table) AddHeaders(headers ...interface{}) {
	t.t.AddHeaders(headers...)
}

// AddRow takes a sequence of cells for one table body row.
func (t *Table) AddRow(cells ...interface{}) {
	t.t.AddRow(cells...)
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
func (t *Table) Render() string {
	return t.t.Render()
}

// Alignment indicates our table column alignments.  We only use per-column
// alignment and do not pass through any per-cell alignments.
type Alignment int

// These constants define how a given column of the table should have each
// cell aligned.
const (
	LEFT Alignment = iota
	CENTER
	RIGHT
)

// AlignColumn sets the alignment of one column in a given table.  It counts
// columns starting from 1.
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
