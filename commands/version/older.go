//go:build !go1.18
// +build !go1.18

package version

import (
	"runtime/debug"
)

func showBuildSettings(buildInfo *debug.BuildInfo) {}
