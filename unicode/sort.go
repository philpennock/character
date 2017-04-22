// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package unicode

import (
	"sort"
)

// CharInfoList is a convenience wrapper for []CharInfo supporting sorting by
// Unicode code-point.
type CharInfoList []CharInfo

func (cil CharInfoList) Len() int           { return len(cil) }
func (cil CharInfoList) Less(i, j int) bool { return cil[i].Number < cil[j].Number }
func (cil CharInfoList) Swap(i, j int)      { cil[i], cil[j] = cil[j], cil[i] }

// Sort sorts a CharInfoList, where sorting is defined as being by Unicode codepoint.
func (cil CharInfoList) Sort() { sort.Sort(cil) }
