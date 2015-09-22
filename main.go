package main

//go:generate ./util/update_entities
//go:generate ./util/update_unicode

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
