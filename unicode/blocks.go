// Copyright © 2015,2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"strings"
)

// What we need is a SegmentTreeMap, such that lookup of a key matches an entry
// where the matcher is a range.
// FIXME: look again for an existing implementation of SegmentTreeMap, or write one.
//
// In the meantime, we know the static data which we're feeding in, and that
// it's less than 300 entries and pre-sorted.  Linear scan will get us working.

// BlockInfo holds the core information to describe a range of Unicode
// characters which make up a Unicode Block.
type BlockInfo struct {
	Min, Max rune
	Name     string
}

// Blocks is our opaque container for holding data to be used for looking up
// block-based information.
type Blocks struct {
	ordered           []BlockInfo
	maxKnownBlockRune rune
}

// Lookup returns the name of the one block which contains a given rune, or
// the empty string if no such block is found.
func (b Blocks) Lookup(r rune) (blockname string) {
	if r > b.maxKnownBlockRune {
		return ""
	}
	for i := range b.ordered {
		if b.ordered[i].Max < r {
			continue
		}
		if b.ordered[i].Min <= r {
			return b.ordered[i].Name
		}
		return "<gap>"
	}
	return ""
}

// FindByName returns the extent of the given block, with start and end runes;
// the block name needs to be "sufficiently unique".
// Returns 0,0,nil if not found.
// The candidateNames []string will be empty unless we hit "insufficiently unique"
func (b Blocks) FindByName(name string) (min, max rune, candidateNames []string) {
	uc := strings.ToUpper(name)
	candidates := make([]BlockInfo, 0, 10)
	for _, block := range b.ordered {
		thisName := strings.ToUpper(block.Name)
		if uc == thisName {
			return block.Min, block.Max, nil
		}
		if strings.Contains(thisName, uc) {
			candidates = append(candidates, block)
		}
	}
	if len(candidates) == 1 {
		return candidates[0].Min, candidates[0].Max, nil
	} else if len(candidates) > 1 {
		candidateNames = make([]string, len(candidates))
		for i := range candidates {
			candidateNames[i] = candidates[i].Name
		}
		return 0, 0, candidateNames
	}
	return 0, 0, nil
}

// ListBlocks returns an ordered list of known blocks.
func (b Blocks) ListBlocks() []BlockInfo {
	return b.ordered
}

// LoadBlocks returns a Blocks holder for BlockInfo lookup
// This is much simpler now that we generate static Golang code for the blocks.
func LoadBlocks() Blocks {
	return Blocks{
		ordered:           allKnownBlocks,
		maxKnownBlockRune: maxKnownBlockRune,
	}
}
