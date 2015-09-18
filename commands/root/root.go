package root

import (
	"sync"

	"github.com/spf13/cobra"
)

var characterCmd = &cobra.Command{
	Use:   "character",
	Short: "character performs character lookups and conversions",
}

var errorCount struct {
	sync.Mutex
	value int
}

func AddCommand(cmds ...*cobra.Command) {
	characterCmd.AddCommand(cmds...)
}

func Errored() {
	errorCount.Lock()
	errorCount.value += 1
	errorCount.Unlock()
}

func GetErrorCount() int {
	errorCount.Lock()
	defer errorCount.Unlock()
	return errorCount.value
}

func Start() {
	characterCmd.Execute()
}
