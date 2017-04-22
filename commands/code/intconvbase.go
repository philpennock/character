// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package code

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/spf13/pflag"
)

// ErrNotValidBaseInt is returned when we're told to interpret a character
// sequence as a number in a base system which we don't support.  We support
// natural number bases between 2 and 32 inclusive, or 0 to imply
// auto-detection via normal Golang rules.
var ErrNotValidBaseInt = errors.New("not a valid numeric base (0 or 2..32)")

// intconvBase is a number value-constrained to be either 0 or 2-32.
type intconvBase int

func (i *intconvBase) String() string {
	return fmt.Sprintf("%d", *i)
}

func (i *intconvBase) Set(nv string) error {
	v, err := strconv.ParseUint(nv, 10, 8)
	if err != nil {
		// should we just return ErrNotValidBaseInt here too?
		return err
	}
	if v == 0 || (2 <= v && v <= 32) {
		*i = intconvBase(v)
		return nil
	}
	return ErrNotValidBaseInt
}

func (i *intconvBase) Int() int { return int(*i) }
func (i *intconvBase) Get() int { return int(*i) }

func (i *intconvBase) Type() string { return "intconvBase" }

// beware the under-documented addition of Type, which just affirms that this
// check is useful.
var _x = intconvBase(0)
var _ pflag.Value = &_x
