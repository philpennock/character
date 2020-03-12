// Copyright © 2015-2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package root

import (
	"fmt"
	"os"
	"runtime/pprof"
	"sync"

	"github.com/mattn/go-shellwords"
	"github.com/spf13/cobra"
)

var globalFlags struct {
	profileCPUFile string
	version        bool
	shellParseArgv string
}

type reExecTrigger struct {
	originalArgs []string
	newArgs      []string
}

func (e reExecTrigger) Error() string { return "re-exec" }

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
		if globalFlags.shellParseArgv != "" {
			// This is so that the WASM version doesn't need JavaScript to
			// understand how to split apart shell quoting.
			newArgs, err := shellwords.Parse(globalFlags.shellParseArgv)
			if err != nil {
				return err
			}
			cmd.SilenceErrors = true
			cmd.SilenceUsage = true
			return reExecTrigger{
				originalArgs: args,
				newArgs:      newArgs,
			}
		}
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if globalFlags.profileCPUFile != "" {
			pprof.StopCPUProfile()
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if globalFlags.version {
			cmd.SetArgs([]string{"version"})
			return cmd.Execute()
		}
		return fmt.Errorf("need a sub-command")
	},
}

func init() {
	// We want to work on flags which must be applied directly to this command,
	// _before_ sub-commands.  Thus "character [--global-flags] subcmd [--cmd-flags]".
	// I can't figure out how to do that with cobra, so for now we have the
	// global-flags applying within sub-commands too.  This is subject to change in
	// the future.
	//
	// If changing this, remember to check other *.go files in this directory
	// for any init()-flag-setting there too.
	flagSet := characterCmd.PersistentFlags()
	flagSet.StringVar(&globalFlags.profileCPUFile, "profile-cpu-file", "", "write CPU profile to file")
	flagSet.BoolVar(&globalFlags.version, "version", false, "alias for version sub-command")
	characterCmd.MarkFlagFilename("profile-cpu-file")
	flagSet.StringVar(&globalFlags.shellParseArgv, "shell-parse-argv", "", "replace argv with shell-split of this string")
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
	err := characterCmd.Execute()
	if err != nil {
		if reExec, ok := err.(reExecTrigger); ok {
			characterCmd.SilenceErrors = false
			characterCmd.SilenceUsage = false
			characterCmd.SetArgs(reExec.newArgs)
			globalFlags.shellParseArgv = ""
			err = characterCmd.Execute()
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "command failed: %s\n", err)
		Errored()
	}
}

// Errorf is a convenience for errors from other commands so that things are consistent
// instead of importing fmt and os all over the place
func Errorf(spec string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, spec, args...)
	Errored()
}

// Cobra exposes the root-level cobra object
func Cobra() *cobra.Command {
	return characterCmd
}
