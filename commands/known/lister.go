// Copyright © 2017 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package known

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/philpennock/character/commands/root"
	"github.com/philpennock/character/internal/table"
)

type lister struct {
	command       string
	nonTableLabel string
	columnTitles  []interface{}
	fieldsExtract func(x interface{}) []interface{}

	nameOnly bool
	verbose  bool
}

func (l *lister) fillDefaults() {
	if l.fieldsExtract == nil {
		l.fieldsExtract = func(x interface{}) []interface{} { return []interface{}{x.(string)} }
	}
}

// Each iterates over a non-nil slice, showing the contents on stdout.
func (l *lister) Each(enumerableI interface{}) {
	l.fillDefaults()

	eVal := reflect.ValueOf(enumerableI)
	if eVal.Kind() != reflect.Slice {
		panic("not given a slice but instead a " + eVal.Kind().String())
	}
	if eVal.IsNil() {
		panic("given a nil slice")
	}
	length := eVal.Len()

	if !l.verbose {
		if !l.nameOnly {
			fmt.Printf("%s %s: ", root.Cobra().Name(), l.command)
			fmt.Printf(l.nonTableLabel, length)
			fmt.Println(":")
		}
		spec := strings.Repeat("\t%q", len(l.columnTitles)) + "\n"
		for i := 0; i < length; i++ {
			item := eVal.Index(i).Interface()
			if l.nameOnly {
				fmt.Printf("%s\n", l.fieldsExtract(item)[0])
			} else {
				fmt.Printf(spec, l.fieldsExtract(item)...)
			}
		}
		return
	}

	if !table.Supported() {
		root.Errorf("sorry, this build is missing table support??\n")
		return
	}

	t := table.New()
	t.AddHeaders(l.columnTitles...)
	for i := 0; i < length; i++ {
		item := eVal.Index(i).Interface()
		t.AddRow(l.fieldsExtract(item)...)
	}
	fmt.Print(t.Render())
}
