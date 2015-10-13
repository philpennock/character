package main

import (
	"testing"

	"github.com/philpennock/character/commands/root"
	"github.com/philpennock/character/sources"

	_ "github.com/philpennock/character/commands/name"
	_ "github.com/philpennock/character/commands/named"
	_ "github.com/philpennock/character/commands/version"
)

const (
	minSaneByRune = 1000
	minSaneByName = 1000
)

func Test000LoadDataExternal(*testing.T) {
	// just load data once so that times for other tests are sane
	_ = sources.NewAll()
}

func TestBasicLoadWithoutErrors(t *testing.T) {
	_ = sources.NewAll()
	if ec := root.GetErrorCount(); ec != 0 {
		t.Fatalf("basic data load gave us %d errors", ec)
	}
}

func TestHaveUnicode(t *testing.T) {
	s := sources.NewAll()
	if runeCount := len(s.Unicode.ByRune); runeCount < minSaneByRune {
		t.Errorf("only got %d entries in unicode ByRune map", runeCount)
	}
	if nameCount := len(s.Unicode.ByName); nameCount < minSaneByName {
		t.Errorf("only got %d entries in unicode ByRune map", nameCount)
	}
}
