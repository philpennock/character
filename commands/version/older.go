// Copyright © 2022 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

//go:build !go1.18

package version

import (
	"runtime/debug"
)

func showBuildSettings(buildInfo *debug.BuildInfo) {}
