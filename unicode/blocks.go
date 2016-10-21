// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"sync"

	"github.com/philpennock/character/aux"
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

func (b Blocks) ListBlocks() []BlockInfo {
	return b.ordered
}

var oneBlocks struct {
	sync.Once
	b Blocks
}

// LoadBlocks returns a Blocks holder for BlockInfo lookup
func LoadBlocks() Blocks {
	oneBlocks.Do(func() {
		oneBlocks.b = parseRawBlocks()
		// NB: at this time, the sorting is done at parse-time with a panic
		// if out-of-order.
	})
	return oneBlocks.b
}

func parseRawBlocks() Blocks {
	b := bytes.NewBuffer(rawBlocks)
	blocks := Blocks{
		ordered:           make([]BlockInfo, 0, rawBlocksLineCount),
		maxKnownBlockRune: 0,
	}

	// ordered []BlockInfo
	// maxKnownBlockRune rune

	matcher := regexp.MustCompile(`^([0-9A-Fa-f]+)\.\.([0-9A-Fa-f]+);\s+(\S.*?)\s*$`)

	lineNum := 0
	for {
		if b.Len() == 0 {
			break
		}
		line, err := b.ReadBytes('\n')
		lineNum++
		if err != nil {
			switch err {
			case io.EOF:
				break
			default:
				panic(err.Error())
			}
		}
		line = line[:len(line)-1]

		// our embedding inserts an extra newline at the start; be resistant
		if len(line) == 0 {
			continue
		}

		got := matcher.FindSubmatch(line)
		if got == nil {
			continue
		}

		bi := BlockInfo{
			Min:  aux.RuneFromHexField(got[1]),
			Max:  aux.RuneFromHexField(got[2]),
			Name: string(got[3]),
		}
		if bi.Max < blocks.maxKnownBlockRune {
			panic(fmt.Sprintf("unsorted block info line %d got max %d which < %d", lineNum, bi.Max, blocks.maxKnownBlockRune))
		}
		// fmt.Printf("Found block %v\n", bi)
		blocks.maxKnownBlockRune = bi.Max
		blocks.ordered = append(blocks.ordered, bi)
	}

	// fmt.Printf("Have %d blocks with max rune %x\n", len(blocks.ordered), blocks.maxKnownBlockRune)
	return blocks
}
