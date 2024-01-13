// Copyright © 2015,2016,2020,2022,2024 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package version

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	json    bool
	verbose bool
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version of character",
	Run: func(cmd *cobra.Command, args []string) {
		if flags.json {
			emitJSONVersionData(cmd.Root().Name())
			return
		}
		showGoModuleVersions(cmd.Root().Name())
		if flags.verbose {
			// keep this after the GoModule one, it mutates the VersionString
			showOldStyleVersions(cmd.Root().Name())
		}

	},
}

func init() {
	versionCmd.Flags().BoolVarP(&flags.verbose, "verbose", "v", false, "show extra stuff")
	versionCmd.Flags().BoolVarP(&flags.json, "json", "", false, "emit JSON version information")
	root.AddCommand(versionCmd)
}

// This JSON data should be usable by tools without those tools breaking for us, ever.
func emitJSONVersionData(programName string) {
	jv := JSONVersion{
		Name:    programName,
		Version: VersionString,
		URL:     SourceURL,
	}
	if len(os.Args) > 0 {
		jv.ArgvInvokedName = filepath.Base(os.Args[0])
	}
	jv.Go.Runtime = runtime.Version()
	jv.Library = make(map[string]LibraryDetailsJSON, len(libraryVersionFuncs))

	for _, f := range libraryVersionFuncs {
		name, infoLines := f()
		jv.Library[name] = LibraryDetailsJSON{Lines: infoLines}
	}
	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		jv.ModuleVersion = make(map[string]string, len(buildInfo.Deps))
		jv.BuildSetting = make(map[string]string, len(buildInfo.Settings))
		for i := range buildInfo.Deps {
			jv.ModuleVersion[buildInfo.Deps[i].Path] = buildInfo.Deps[i].Version
		}
		for i := range buildInfo.Settings {
			jv.BuildSetting[buildInfo.Settings[i].Key] = buildInfo.Settings[i].Value
		}
	}

	jout := json.NewEncoder(os.Stdout)
	jout.SetIndent("", "  ")
	if err := jout.Encode(jv); err != nil {
		panic(err.Error())
	}
}

// JSONVersion provides a publicly guaranteed backwards-compatible
// machine-parseable JSON version.  Fields can be added, but must not be removed,
// and must not be frivolously switched to be empty.
type JSONVersion struct {
	Version         string `json:"version"`
	Name            string `json:"name"`
	ArgvInvokedName string `json:"argv_invoked_name"`
	URL             string `json:"url"`

	Go struct {
		Runtime string `json:"runtime"`
	} `json:"go"`

	// The Library is keyed by the name of a library and each item includes
	// unstructed lines of version data from that library; we wrap it in an
	// extra layer so that if there's more per-library structure _available_
	// for some libraries in future, we don't need to restructure our output.
	//
	// Theoretically two functions might report the same name, but since that's
	// from our library handling, that would be our mistake.
	Library map[string]LibraryDetailsJSON `json:"library"`

	ModuleVersion map[string]string `json:"module_version"`
	BuildSetting  map[string]string `json:"build_setting"`
}

type LibraryDetailsJSON struct {
	Lines []string `json:"lines"`
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
		fmt.Print(t.Render())
		showBuildSettings(buildInfo)
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
