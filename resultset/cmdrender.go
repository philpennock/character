// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package resultset

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

// Rather than repeat verbose/net-verbose/etc in each command, have flags here

var resultCmdFlags struct {
	internalDebug bool
	oneline       bool
	netVerbose    bool
	verbose       bool
}

// RegisterCmdFlags adds the --verbose/--net-verbose/--internal-debug flags to a Cobra cmd.
func RegisterCmdFlags(cmd *cobra.Command, supportOneline bool) {
	if supportOneline {
		cmd.Flags().BoolVarP(&resultCmdFlags.oneline, "oneline", "1", false, "multiple chars on one line")
	}
	if !CanTable() {
		return
	}
	cmd.Flags().BoolVarP(&resultCmdFlags.internalDebug, "internal-debug", "", false, "")
	cmd.Flags().MarkHidden("internal-debug")
	cmd.Flags().BoolVarP(&resultCmdFlags.netVerbose, "net-verbose", "N", false, "show net-biased information (punycode, etc)")
	cmd.Flags().BoolVarP(&resultCmdFlags.verbose, "verbose", "v", false, "show information about the character")
}

// CmdVerbose indicates if a documented verbose flag was set; occasionally commands want to know
// for reasons other than a resultset (eg, 'character browse -vB').
func CmdVerbose() bool {
	return resultCmdFlags.netVerbose || resultCmdFlags.verbose
}

// ErrIncompatibleFlags indicates that multiple types of verboseness were simultaneously requested.
var ErrIncompatibleFlags = errors.New("incompatible table-rendering flags")

// FlagsOkay returns either ErrIncompatibleFlags or nil
func FlagsOkay() error {
	c := 0
	if resultCmdFlags.internalDebug {
		c++
	}
	if resultCmdFlags.netVerbose {
		c++
	}
	if resultCmdFlags.verbose {
		c++
	}
	if resultCmdFlags.oneline {
		c++
	}
	if c > 1 {
		return ErrIncompatibleFlags
	}
	return nil
}

// RenderPerCmdline performs the table rendering requested from the
// command-line, or iteration in non-table form, printing the item passed.
// This is otherwise per-command boilerplate.
func (rs *ResultSet) RenderPerCmdline(defaultPI printItem) {
	if resultCmdFlags.verbose {
		rs.PrintTables()
	} else if resultCmdFlags.netVerbose {
		rs.SelectFieldsNet()
		rs.PrintTables()
	} else if resultCmdFlags.internalDebug {
		rs.SelectFieldsDebug()
		rs.PrintTables()
	} else if resultCmdFlags.oneline {
		fmt.Println(rs.StringPlain(defaultPI))
	} else {
		rs.PrintPlain(defaultPI)
	}
}
