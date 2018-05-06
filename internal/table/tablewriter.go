// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// +build tablewriter

/*
This implementation uses tablewriter for layout;
this is nice, but has some issues:

* Alignment is only per-table, not per-column
* Auto-detected alignment differs for decimal numbers from anything else, so a
  column of hex values will have varying alignments, unless we coerce
  everything in one way; this is a limitation of per-table alignment
* Limited styling doesn't provide enough access points for UTF-8 boxes
* No in-table separator rows

The tablewriter package only takes strings; our wrapper at present only needs
to support strings in, so inserting any other type will result in replacement
with "??".  We collect incidences of this into an errCount but at present
never look at it; it's present as a hook for future debugging extension.
*/
package table

import (
	"bytes"

	"github.com/olekukonko/tablewriter"
)

// Supported indicates that we have a terminal table provider loaded.
// This exists in anticipation of multiple providers and tables being optional
// if not supported at all.
func Supported() bool { return true }

// A Table encapsulates our terminal table object from the dependency.
type Table struct {
	t        *tablewriter.Table
	output   *bytes.Buffer
	rowCount int
	errCount int
}

// New gives us a new empty table, configured for our basic requirements.
func New() *Table {
	ours := &Table{
		output: &bytes.Buffer{},
	}
	ours.t = tablewriter.NewWriter(ours.output)
	// Here is where we might override the table styling for UTF-8 support
	// At present, we get .SetColumnSeparator, .SetRowSeparator and
	// .SetCenterSeparator(), which is not enough to handle outside corners
	ours.t.SetAutoWrapText(false)
	return ours
}

// AddHeaders takes a sequence of header-names for each column, and configures
// them as the header row.
func (t *Table) AddHeaders(headers ...interface{}) {
	strHeaders := make([]string, len(headers))
	var ok bool
	for i := range headers {
		strHeaders[i], ok = headers[i].(string)
		if !ok {
			strHeaders[i] = "??"
			t.errCount++
		}
	}
	t.t.SetHeader(strHeaders)
}

// AddRow takes a sequence of cells for one table body row.
func (t *Table) AddRow(cells ...interface{}) {
	row := make([]string, len(cells))
	var ok bool
	for i := range cells {
		row[i], ok = cells[i].(string)
		if !ok {
			row[i] = "??"
			t.errCount++
		}
	}
	t.t.Append(row)
	t.rowCount++
}

// AddSeparator adds one row to the table, as a separator.
func (t *Table) AddSeparator() {
	// no separator support, so skip?
}

// Empty is a predicate indicating if any real content has been added.
func (t *Table) Empty() bool {
	return t.rowCount == 0
}

// Render gives us the full table as a string for display.
func (t *Table) Render() string {
	t.t.Render()
	return t.output.String()
}

// AlignColumn sets the alignment of one column in a given table.  It counts
// columns starting from 1.  In this implementation, it has no effect because
// tablewriter only supports per-table, not per-column, alignments.
func (t *Table) AlignColumn(column int, align Alignment) {
	// IGNORED
	_ = column
	_ = align
}

// SetSkipableColumn sets a column as skipable in some contexts (typically if
// every entry is empty).  In this implementation, it is ignored.
func (t *Table) SetSkipableColumn(column int) {
	_ = column
}
