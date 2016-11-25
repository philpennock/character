// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// +build tabular !tablewriter,!termtables

package version

import "github.com/PennockTech/tabular"

func init() {
	addLibraryVersionFunc(func() (string, []string) {
		// "tabular: " already in tabular.Versions() output,
		// so empty first return value.
		return "", tabular.Versions()
	})
}
