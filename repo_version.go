// Copyright © 2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package main

import "github.com/philpennock/character/commands/version"

// This should be "latest release + 0.0.1 with -devel suffix"
// OR if at the point where we're tagging, then that explicit tag.
const HARDCODED_VERSION = "v0.7.0"

func init() {
	if version.VersionString == "" {
		version.VersionString = HARDCODED_VERSION
	}
}
