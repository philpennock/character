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

var ResultCmdFlags struct {
	internalDebug bool
	Oneline       bool
	NetVerbose    bool
	Verbose       bool
	Emoji         bool
	Text          bool
}

// RegisterCmdFlags adds the --verbose/--net-verbose/--internal-debug flags to a Cobra cmd.
func RegisterCmdFlags(cmd *cobra.Command, supportOneline bool) {
	if supportOneline {
		cmd.Flags().BoolVarP(&ResultCmdFlags.Oneline, "oneline", "1", false, "multiple chars on one line")
	}
	if !CanTable() {
		return
	}
	cmd.Flags().BoolVarP(&ResultCmdFlags.internalDebug, "internal-debug", "", false, "")
	cmd.Flags().MarkHidden("internal-debug")
	cmd.Flags().BoolVarP(&ResultCmdFlags.NetVerbose, "net-verbose", "N", false, "show net-biased information (punycode, etc)")
	cmd.Flags().BoolVarP(&ResultCmdFlags.Verbose, "verbose", "v", false, "show information about the character")
	cmd.Flags().BoolVarP(&ResultCmdFlags.Emoji, "emoji-presentation", "E", false, "force emoji presentation")
	cmd.Flags().BoolVarP(&ResultCmdFlags.Text, "text-presentation", "T", false, "force text presentation")
}

// CmdVerbose indicates if a documented verbose flag was set; occasionally commands want to know
// for reasons other than a resultset (eg, 'character browse -vB').
func CmdVerbose() bool {
	return ResultCmdFlags.NetVerbose || ResultCmdFlags.Verbose
}

// ErrIncompatibleFlags indicates that multiple types of verboseness were simultaneously requested.
var ErrIncompatibleFlags = errors.New("incompatible table-rendering flags")

// FlagsOkay returns either ErrIncompatibleFlags or nil
func FlagsOkay() error {
	c := 0
	if ResultCmdFlags.internalDebug {
		c++
	}
	if ResultCmdFlags.NetVerbose {
		c++
	}
	if ResultCmdFlags.Verbose {
		c++
	}
	if ResultCmdFlags.Oneline {
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
	if ResultCmdFlags.Emoji && defaultPI == PRINT_RUNE {
		defaultPI = PRINT_RUNE_PRESENT_EMOJI
	}
	if ResultCmdFlags.Text && defaultPI == PRINT_RUNE {
		defaultPI = PRINT_RUNE_PRESENT_TEXT
	}
	if ResultCmdFlags.Verbose {
		rs.PrintTables()
	} else if ResultCmdFlags.NetVerbose {
		rs.SelectFieldsNet()
		rs.PrintTables()
	} else if ResultCmdFlags.internalDebug {
		rs.SelectFieldsDebug()
		rs.PrintTables()
	} else if ResultCmdFlags.Oneline {
		fmt.Println(rs.StringPlain(defaultPI))
	} else {
		rs.PrintPlain(defaultPI)
	}
}
