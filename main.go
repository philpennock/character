// Copyright © 2015,2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package main

//go:generate go run ./util/update_entities.go
//go:generate go run ./util/update_unicode.go
//go:generate ./util/update_static_vim
//go:generate go run ./util/update_x11_compose.go

import (
	"os"

	"github.com/philpennock/character/commands/root"

	_ "github.com/philpennock/character/commands/browse"
	_ "github.com/philpennock/character/commands/code"
	_ "github.com/philpennock/character/commands/known"
	_ "github.com/philpennock/character/commands/name"
	_ "github.com/philpennock/character/commands/named"
	_ "github.com/philpennock/character/commands/puny"
	_ "github.com/philpennock/character/commands/region"
	_ "github.com/philpennock/character/commands/transform"
	_ "github.com/philpennock/character/commands/version"

	_ "github.com/philpennock/character/commands/deprecated"
)

func main() {
	root.Start()
	if root.GetErrorCount() > 0 {
		os.Exit(1)
	}
}
