// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package sources

import (
	"github.com/philpennock/character/unicode"
)

// Sources encapsulates all the sources of data we use for various lookups.
type Sources struct {
	Unicode unicode.Unicode
	UBlocks unicode.Blocks
	Vim     VimData
}

// NewEmpty gives us a Sources item with no data loaded.
func NewEmpty() *Sources {
	return &Sources{}
}

// LoadUnicode gives us static information about Unicode data sources.
func (s *Sources) LoadUnicode() *Sources {
	s.Unicode = unicode.Load()
	return s
}

// LoadUnicodeSearch gives us static information about Unicode data sources,
// but also gives us substring search capabilities.
func (s *Sources) LoadUnicodeSearch() *Sources {
	s.Unicode = unicode.LoadSearch()
	return s
}

// LoadUnicodeBlocks makes available Unicode block information
func (s *Sources) LoadUnicodeBlocks() *Sources {
	s.UBlocks = unicode.LoadBlocks()
	return s
}

// LoadLiveVim gives us dynamically-loaded data about vim digraphs, retrieved
// by invoking vim.  Will handle vim not being installed (but will print
// errors; we'll probably change this in the future).
func (s *Sources) LoadLiveVim() *Sources {
	s.Vim = loadVimDigraphsCached()
	return s
}

// LoadLiveVimAgain avoids the cache so that data is loaded from Vim again
func (s *Sources) LoadLiveVimAgain() *Sources {
	s.Vim = loadVimDigraphs()
	return s
}

// LoadStaticVim gives us vim digraphs built into character, based upon
// shipping digraphs as of some unspecified version of vim.
func (s *Sources) LoadStaticVim() *Sources {
	s.Vim = loadStaticVimDigraphs()
	return s
}

// NewFast gives us a Sources item which has the fast data; no search, no live vim
func NewFast() *Sources {
	return NewEmpty().LoadUnicode().LoadUnicodeBlocks().LoadStaticVim()
}

// NewAll gives us a Sources item which has had loaded all the data sources
// that we know about.
func NewAll() *Sources {
	return NewEmpty().LoadUnicodeSearch().LoadLiveVim()
}
