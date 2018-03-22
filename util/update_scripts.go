// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// +build ignore

// Domain registration requires that domains have characters from consistent
// scripts; zones beneath the root zone should only accept registrations from
// one of a specific named list of scripts, and a given domain must be within
// one script.
//
// ICANN maintain the definition of what's in a script.
// Per <https://www.icann.org/en/system/files/files/msr-2-overview-14apr15-en.pdf>
// the current list is <https://www.icann.org/en/system/files/files/msr-2-wle-rules-13apr15-en.xml>
//
// The current definition should be downloaded into `unicode/msr-current-rules.xml`
// and then this script run.

package main

type XMLLGR struct {
	Meta *XMLLGRMeta `xml:meta`
	Data *XMLLGRData `xml:data`
	Rules *XMLLGRRules `xml:rules`
}

type XMLLGRMeta struct {
}

type XMLLGRData struct {
	// Array of items, each either `range` or `char`
}

type XMLLGRRules struct {
	Rules []XMLLGRRule
	Actions []XMLLGRAction
}

type XMLLGRRule struct {
}

type XMLLGRAction struct {
	// disp; any-variant; match
}


import (
)
