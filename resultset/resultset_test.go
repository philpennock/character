// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package resultset

import (
	"bytes"
	"errors"
	"testing"

	"github.com/liquidgecka/testlib"

	"github.com/philpennock/character/sources"
	//"github.com/philpennock/character/table"
	"github.com/philpennock/character/unicode"
)

// We currently don't want to test external data loading for resultsets, we can
// test that elsewhere.  So skip to .NewFast().  If we change our minds,
// then rename this back to start with Test.
func SkipMeTest000LoadDataExternal(*testing.T) {
	// just load data once so that times for other tests are sane
	_ = sources.NewAll()
}

func TestHaveTableSupport(t *testing.T) {
	if !CanTable() {
		t.Error("missing Table support")
	}
}

func TestResultSetBasics(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()

	srcs := sources.NewFast()
	rs := New(srcs, 10)

	T.Equal(rs.ErrorCount(), 0, "no errors in fresh resultset")

	ci := unicode.CharInfo{
		Number: '✓',
		Name:   "CHECK MARK",
	}

	b := new(bytes.Buffer)
	rs.OutputStream = b
	rs.ErrorStream = b

	shouldBe := func(desired, msg string, reset bool) {
		if reset {
			b.Reset()
		}
		rs.PrintPlain(PRINT_RUNE)
		have := b.String()
		T.Equal(have, desired, msg)
	}

	rs.AddCharInfo(ci)
	rs.AddCharInfo(ci)
	shouldBe("✓\n✓\n", "printed two check marks", true)
	shouldBe("✓\n✓\n✓\n✓\n", "printed another two check marks", false)
	rs.AddCharInfo(ci)
	shouldBe("✓\n✓\n✓\n", "printed three check marks", true)
	rs.AddDivider()
	shouldBe("✓\n✓\n✓\n\n", "printed three check marks and divider", true)
	rs.AddCharInfo(ci)
	shouldBe("✓\n✓\n✓\n\n✓\n", "printed some check marks with divider in there", true)

	rs.AddError("dummy", errors.New("pseudo-error goes here"))
	rs.AddCharInfo(ci)
	shouldBe("✓\n✓\n✓\n\n✓\nlooking up \"dummy\": pseudo-error goes here\n✓\n", "printed some check-marks and an error", true)
}

func TestBlockNames(t *testing.T) {
	T := testlib.NewT(t)
	defer T.Finish()
	srcs := sources.NewFast()
	rs := New(srcs, 10)

	for _, pair := range []struct {
		r  rune
		bn string
	}{
		{'£', "Latin-1 Supplement"},
		{'€', "Currency Symbols"},
		{'→', "Arrows"},
		{'♡', "Miscellaneous Symbols"},
		{'τ', "Greek and Coptic"},
		{'÷', "Latin-1 Supplement"},
		{'†', "General Punctuation"},
		{'🌸', "Miscellaneous Symbols and Pictographs"},
		{'x', "Basic Latin"},
	} {
		ci := charItem{unicode: unicode.CharInfo{Number: pair.r}}
		have := rs.RenderCharInfoItem(ci, PRINT_BLOCK)
		T.Equal(have, pair.bn, "block lookup of known char should be known")
	}
}
