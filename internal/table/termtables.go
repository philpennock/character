// Copyright © 2015,2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// We want `termtables` to activate us, and to be the default if no other
// implementation is provided
//
// 2020: apcera/termtables has gone, github.com/xlab/tablewriter is a
// widespread fork which ripped out the terminal querying and locale handling.
// Not bumping the copyright year for replacing one import path string.

//go:build termtables
// +build termtables

package table

import (
	"github.com/xlab/tablewriter"
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

// SetSkipableColumn sets a column as skipable in some contexts (typically if
// every entry is empty).  In this implementation, it is ignored.
func (t *Table) SetSkipableColumn(column int) {
	_ = column
}
