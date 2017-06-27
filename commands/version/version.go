// Copyright © 2015,2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
)

// SourceURL is so that the version command identifies where this came from.
// It can be overriden at link time, but is not expected to be.
var SourceURL = "https://github.com/philpennock/character"

// VersionString is expected to be set by the linker during build.
// If make(1) is used for build, this will happen.
var VersionString string

// Library version functions should be defined in other files in this dir,
// under appropriate build-tag constraints, and those files should use
// the add function below in their init() routines.  Each function should
// return a short library name, and any lines of output for versioning.
// If the short library name is empty, then the library's name is assumed
// to be embedded in the output lines already.
var libraryVersionFuncs []func() (string, []string)

func addLibraryVersionFunc(f func() (string, []string)) {
	libraryVersionFuncs = append(libraryVersionFuncs, f)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version of character",
	Run: func(cmd *cobra.Command, args []string) {
		if VersionString == "" {
			VersionString = "<unknown>"
		}
		fmt.Printf("%s: version %s\n", cmd.Root().Name(), VersionString)
		fmt.Printf("Golang: Runtime: %s\n", runtime.Version())
		for _, f := range libraryVersionFuncs {
			name, infoLines := f()
			for _, l := range infoLines {
				if name != "" {
					fmt.Printf("%s: %s\n", name, l)
				} else {
					fmt.Printf("%s\n", l)
				}
			}
		}
		fmt.Printf("%s: Source URL <%s>\n", cmd.Root().Name(), SourceURL)
	},
}

func init() {
	root.AddCommand(versionCmd)
}
