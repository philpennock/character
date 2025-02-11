// Copyright © 2022 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

//go:build go1.18
// +build go1.18

package version

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/philpennock/character/internal/table"
)

func showBuildSettings(buildInfo *debug.BuildInfo) {
	t := table.New()
	t.AddHeaders("Setting", "Value")
	var k, v string
	for i := range buildInfo.Settings {
		k = buildInfo.Settings[i].Key
		v = buildInfo.Settings[i].Value
		switch k {
		case "-ldflags":
			// not quite safe against -X inside quoted strings but close enough for our purposes
			v = strings.Replace(v, " -X ", "\n-X ", -1)
		case "DefaultGODEBUG":
			v = strings.Replace(v, ",", ",\n", -1)
		}
		t.AddRow(k, v)
	}
	fmt.Print(t.Render())
}
