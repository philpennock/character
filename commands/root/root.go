// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package root

import (
	"os"
	"runtime/pprof"
	"sync"

	"github.com/spf13/cobra"
)

var globalFlags struct {
	profileCPUFile string
}

var characterCmd = &cobra.Command{
	Use:   "character",
	Short: "character performs character lookups and conversions",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if globalFlags.profileCPUFile != "" {
			f, err := os.Create(globalFlags.profileCPUFile)
			if err != nil {
				return err
			}
			pprof.StartCPUProfile(f)
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if globalFlags.profileCPUFile != "" {
			pprof.StopCPUProfile()
		}
	},
}

func init() {
	characterCmd.PersistentFlags().StringVar(&globalFlags.profileCPUFile, "profile-cpu-file", "", "write CPU profile to file")
}

var errorCount struct {
	sync.Mutex
	value int
}

// AddCommand is the hook used by our sub-commands to register themselves
// into our CLI dispatch system.  Per-module init() hooks should use this.
func AddCommand(cmds ...*cobra.Command) {
	characterCmd.AddCommand(cmds...)
}

// Errored bumps an error count, used to determine if the main program should
// exit false or not.
func Errored() {
	errorCount.Lock()
	errorCount.value++
	errorCount.Unlock()
}

// GetErrorCount is intended for use by main(), to determine how to exit.
func GetErrorCount() int {
	errorCount.Lock()
	defer errorCount.Unlock()
	return errorCount.value
}

// Start is the entry-point used by main(), after all the sub-modules have
// registered their availability via AddCommand calls in their init functions.
func Start() {
	characterCmd.Execute()
}
