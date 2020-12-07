// Copyright © 2017,2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package resultset

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Rather than repeat verbose/net-verbose/etc in each command, have flags here

var ResultCmdFlags struct {
	internalDebug   bool
	implicitVerbose bool
	JSON            bool
	Oneline         bool
	NetVerbose      bool
	Verbose         bool
	Emoji           bool
	Text            bool
	Left            bool
	Right           bool
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
	cmd.Flags().BoolVarP(&ResultCmdFlags.JSON, "json", "J", false, "show JSON output")
	cmd.Flags().BoolVarP(&ResultCmdFlags.NetVerbose, "net-verbose", "N", false, "show net-biased information (punycode, etc)")
	cmd.Flags().BoolVarP(&ResultCmdFlags.Verbose, "verbose", "v", false, "show information about the character")
	cmd.Flags().BoolVarP(&ResultCmdFlags.Emoji, "emoji-presentation", "E", false, "force emoji presentation")
	cmd.Flags().BoolVarP(&ResultCmdFlags.Text, "text-presentation", "T", false, "force text presentation")
	cmd.Flags().BoolVarP(&ResultCmdFlags.Left, "left", "L", false, "emoji facing left")
	cmd.Flags().BoolVarP(&ResultCmdFlags.Right, "right", "R", false, "emoji facing right")

	cmd.Flags().BoolVarP(&ResultCmdFlags.implicitVerbose, "internal-implicit-verbose", "", false, "")
	cmd.Flags().MarkHidden("internal-implicit-verbose")
}

// CmdVerbose indicates if a documented verbose flag was set; occasionally commands want to know
// for reasons other than a resultset (eg, 'character browse -vB').
func CmdVerbose() bool {
	return ResultCmdFlags.NetVerbose || ResultCmdFlags.Verbose
}

// ErrIncompatibleFlags indicates that multiple types of verboseness were simultaneously requested.
type ErrIncompatibleFlags []string

func (e ErrIncompatibleFlags) Error() string {
	return fmt.Sprintf("incompatible table-rendering flags: %v", []string(e))
}

// FlagsOkay returns either ErrIncompatibleFlags or nil
func FlagsOkay() error {
	onlyOne := make([]string, 0, 5)
	if ResultCmdFlags.internalDebug {
		onlyOne = append(onlyOne, "--internal-debug")
	}
	if ResultCmdFlags.NetVerbose {
		onlyOne = append(onlyOne, "--net-verbose|-N")
	}
	if ResultCmdFlags.Verbose {
		onlyOne = append(onlyOne, "--verbose|-v")
	}
	if ResultCmdFlags.Oneline {
		onlyOne = append(onlyOne, "--oneline|-1")
	}
	if ResultCmdFlags.JSON {
		onlyOne = append(onlyOne, "--json|-J")
	}
	if len(onlyOne) > 1 {
		return ErrIncompatibleFlags(onlyOne)
	}
	if len(onlyOne) == 0 && ResultCmdFlags.implicitVerbose {
		ResultCmdFlags.Verbose = true
	}

	// The left/right direction indicator sequences end with the emoji
	// presentation selector, so emoji is implicit for those, and not
	// incompatible; we do need to ensure we're not pre-empted.
	onlyOne = make([]string, 0, 5)
	if ResultCmdFlags.Left {
		onlyOne = append(onlyOne, "--left|-L")
		ResultCmdFlags.Emoji = false // implicit
	}
	if ResultCmdFlags.Right {
		onlyOne = append(onlyOne, "--right|-R")
		ResultCmdFlags.Emoji = false // implicit
	}
	if ResultCmdFlags.Text {
		onlyOne = append(onlyOne, "--text|-T")
	}
	if ResultCmdFlags.Emoji && ResultCmdFlags.Text {
		onlyOne = append(onlyOne, "--emoji|-E")
	}
	if len(onlyOne) > 1 {
		return ErrIncompatibleFlags(onlyOne)
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
	if ResultCmdFlags.Left && defaultPI == PRINT_RUNE {
		defaultPI = PRINT_RUNE_PRESENT_LEFT
	}
	if ResultCmdFlags.Right && defaultPI == PRINT_RUNE {
		defaultPI = PRINT_RUNE_PRESENT_RIGHT
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
	} else if ResultCmdFlags.JSON {
		rs.PrintJSON()
	} else {
		rs.PrintPlain(defaultPI)
	}
}
