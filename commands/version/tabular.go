// Copyright © 2016,2024 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

//go:build tabular || (!tablewriter && !termtables)
// +build tabular !tablewriter,!termtables

package version

import (
	"strings"

	"go.pennock.tech/tabular"
)

func init() {
	addLibraryVersionFunc(func() (string, []string) {
		// "tabular: " already in tabular.Versions() output,
		// but for the JSON output we need the first string to be non-empty, so
		// we strip it if present.
		const libraryName = "tabular"
		const autostrip = "tabular: "
		const stripLen = len(autostrip)
		lines := tabular.Versions()
		for i := range lines {
			if strings.HasPrefix(lines[i], autostrip) {
				lines[i] = lines[i][stripLen:]
			}
		}
		return libraryName, lines
	})
}
