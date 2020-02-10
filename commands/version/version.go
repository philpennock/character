// Copyright © 2015,2016,2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package version

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/philpennock/character/commands/root"
	"github.com/philpennock/character/internal/table"
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

var flags struct {
	verbose bool
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version of character",
	Run: func(cmd *cobra.Command, args []string) {
		showGoModuleVersions(cmd.Root().Name())
		if flags.verbose {
			// keep this after the GoModule one, it mutates the VersionString
			showOldStyleVersions(cmd.Root().Name())
		}

	},
}

func init() {
	versionCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show extra stuff")
	root.AddCommand(versionCmd)
}

func showGoModuleVersions(programName string) {
	// It amuses me that our existing --table-style top-level option comes
	// along for free and this can be HTML, JSON, whatever.  Just need to
	// junk all of the above.
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		t := table.New()
		t.AddHeaders("Module Path", "Version", "Sum", "Replaced")
		m := &buildInfo.Main

		topVersion := m.Version
		if VersionString != "" {
			topVersion = VersionString
		}
		t.AddRow(m.Path, topVersion, m.Sum, m.Replace != nil)

		for m.Replace != nil {
			m = m.Replace
			t.AddRow(m.Path, m.Version, m.Sum, m.Replace != nil)
		}
		for _, m := range buildInfo.Deps {
			t.AddRow(m.Path, m.Version, m.Sum, m.Replace != nil)
			for m.Replace != nil {
				m = m.Replace
				t.AddRow(m.Path, m.Version, m.Sum, m.Replace != nil)
			}
		}
		// sigh, if I junk the other table interfaces, I can add RenderTo instead of using this.
		fmt.Printf(t.Render())
	}
}

func showOldStyleVersions(programName string) {
	vs := VersionString
	if vs == "" {
		vs = "<unknown>"
	}
	fmt.Printf("\n%s: version %s\n", programName, vs)

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

	fmt.Printf("%s: Source URL <%s>\n", programName, SourceURL)
}
