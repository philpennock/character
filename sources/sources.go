package sources

import (
	"github.com/philpennock/character/unicode"
)

type Sources struct {
	Unicode unicode.Unicode
	Vim     VimData
}

func NewEmpty() *Sources {
	return &Sources{}
}

func (s *Sources) LoadUnicode() *Sources {
	s.Unicode = unicode.Load()
	return s
}

func (s *Sources) LoadVim() *Sources {
	s.Vim = loadVimDigraphs()
	return s
}

func NewAll() *Sources {
	return NewEmpty().LoadUnicode().LoadVim()
}
