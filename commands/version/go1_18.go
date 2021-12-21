//go:build go1.18
// +build go1.18

package version

import (
	"fmt"
	"runtime/debug"

	"github.com/philpennock/character/internal/table"
)

func showBuildSettings(buildInfo *debug.BuildInfo) {
	t := table.New()
	t.AddHeaders("Setting", "Value")
	for i := range buildInfo.Settings {
		t.AddRow(buildInfo.Settings[i].Key, buildInfo.Settings[i].Value)
	}
	fmt.Print(t.Render())
}
