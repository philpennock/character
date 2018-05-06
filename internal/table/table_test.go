// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package table

import (
	"testing"

	"github.com/liquidgecka/testlib"
)

// should this instead be in a sub-package importing the parent, and using
// Example functions with output in comments?
func TestBasicTables(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()
	T.Equal(Supported(), true, "table should be supported")

	// beware https://github.com/apcera/termtables/issues/28 and problems rendering twice if have headers.

	tb := New()
	T.Equal(tb.Empty(), true, "new table should be empty")

	should := "" +
		"┏┓\n" +
		"┗┛\n" +
		""
	T.Equal(tb.Render(), should, "empty table renders")

	tb = New()
	should = "" +
		"┏━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━┓\n" +
		"┃ foo    ┃ loquacious ┃ x    ┃\n" +
		"┣━━━━━━━━╇━━━━━━━━━━━━╇━━━━━━┫\n" +
		"┃ 42     │ .          │ fred ┃\n" +
		"┃ snerty │ word       │ r    ┃\n" +
		"┗━━━━━━━━┷━━━━━━━━━━━━┷━━━━━━┛\n" +
		""
	tb.AddHeaders("foo", "loquacious", "x")
	tb.AddRow(42, ".", "fred")
	tb.AddRow("snerty", "word", "r")
	T.Equal(tb.Render(), should, "basic table should render cleanly")

	tb = New()
	should = "" +
		"┏━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━┓\n" +
		"┃ foo    ┃ loquacious ┃ x        ┃\n" +
		"┣━━━━━━━━╇━━━━━━━━━━━━╇━━━━━━━━━━┫\n" +
		"┃ 42     │ .          │ frederic ┃\n" +
		"┠────────┼────────────┼──────────┨\n" +
		"┃ snerty │ word       │ 3        ┃\n" +
		"┗━━━━━━━━┷━━━━━━━━━━━━┷━━━━━━━━━━┛\n" +
		""
	tb.AddHeaders("foo", "loquacious", "x")
	tb.SetSkipableColumn(2)
	tb.AddRow(42, ".", "frederic")
	tb.AddSeparator()
	tb.AddRow("snerty", "word", 3)
	T.Equal(tb.Render(), should, "basic table with separator added")

	tb = New()
	should = "" +
		"┏━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━┓\n" +
		"┃ foo    ┃ loquacious ┃        x ┃\n" +
		"┣━━━━━━━━╇━━━━━━━━━━━━╇━━━━━━━━━━┫\n" +
		"┃ 42     │ .          │ frederic ┃\n" +
		"┃ snerty │ word       │        3 ┃\n" +
		"┗━━━━━━━━┷━━━━━━━━━━━━┷━━━━━━━━━━┛\n" +
		""
	tb.AddHeaders("foo", "loquacious", "x")
	tb.AddRow(42, ".", "frederic")
	tb.AddRow("snerty", "word", 3)
	tb.AlignColumn(3, RIGHT)
	T.Equal(tb.Render(), should, "basic table with 3rd column right-aligned")
}
