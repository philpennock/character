// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package main

//go:generate ./util/update_entities
//go:generate ./util/update_unicode
//go:generate ./util/update_static_vim

import (
	"os"

	"github.com/philpennock/character/commands/root"

	_ "github.com/philpennock/character/commands/name"
	_ "github.com/philpennock/character/commands/named"
	_ "github.com/philpennock/character/commands/version"
)

func main() {
	root.Start()
	if root.GetErrorCount() > 0 {
		os.Exit(1)
	}
}
