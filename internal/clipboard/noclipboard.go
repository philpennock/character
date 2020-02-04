// Copyright © 2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// +build noclipboard

package clipboard

import "errors"

// There is no clipboard package support.

// Should be okay to switch from var to const?
const Unsupported bool = true

func WriteAll(text string) error { return errors.New("no clipboard support") }
