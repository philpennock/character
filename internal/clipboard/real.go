// Copyright © 2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// +build !noclipboard

package clipboard

import (
	target "github.com/atotto/clipboard"
)

// We use just one function and one variable.
// That's one less of each than atotto's package exports.  :)
//
// We don't want to support looking at the clipboard.
// Fortunately the variable we reference is set during package init, so we can
// just rely upon init ordering and copy it.

var Unsupported bool

func init() {
	Unsupported = target.Unsupported
}

func WriteAll(text string) error { return target.WriteAll(text) }
