package sources

import (
	"github.com/philpennock/character/unicode"
)

// Sources encapsulates all the sources of data we use for various lookups.
type Sources struct {
	Unicode unicode.Unicode
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

// LoadVim gives us dynamically-loaded data about vim digraphs, retrieved
// by invoking vim.  Will handle vim not being installed (but will print
// errors; we'll probably change this in the future).
func (s *Sources) LoadVim() *Sources {
	s.Vim = loadVimDigraphs()
	return s
}

// NewAll gives us a Sources item which has had loaded all the data sources
// that we know about.
func NewAll() *Sources {
	return NewEmpty().LoadUnicode().LoadVim()
}
