package main

import (
	"os"

	"github.com/philpennock/character/commands/root"

	_ "github.com/philpennock/character/commands/named"
)

func main() {
	root.Start()
	if root.GetErrorCount() > 0 {
		os.Exit(1)
	}
}
