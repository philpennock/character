package sources

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unicode/utf8"
)

// VimDigraph encapsulates a vim digraph sequence, which is an input tuple of
// keypresses which together result in a rune.  After typing "Ctrl-K", you
// might enter "Pd" to get "£".
type VimDigraph struct {
	Sequence  string
	Result    rune
	Codepoint int
}

// VimData is the set of all data we have retrieved about characters from vim.
type VimData struct {
	DigraphByRune map[rune][]VimDigraph
}

var cachedVimData struct {
	sync.Once
	d VimData
}

func loadVimDigraphsCached() VimData {
	cachedVimData.Do(func() {
		cachedVimData.d = loadVimDigraphs()
	})
	return cachedVimData.d
}

func loadVimDigraphs() VimData {
	// To make vim work here, note that we have to plumb stdin though,
	// else it exits immediately.  Once we do plumb through, +quit causes
	// vim to exit 1.
	c := exec.Command("vim", "-e", "+digraphs", "+quit")
	c.Stdin = os.Stdin
	b := &bytes.Buffer{}
	c.Stdout = b
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		// this is expected, as long as it's 1 as an exit
		switch err.(type) {
		case *exec.ExitError:
			if e2, ok := err.(*exec.ExitError).Sys().(syscall.WaitStatus); ok {
				// We expect +quit to result in 1
				if e2.ExitStatus() != 1 {
					fmt.Fprintf(os.Stderr, "error getting digraphs from vim: %s\n", err)
					return VimData{}
				}
			} else {
				fmt.Fprintf(os.Stderr, "failed to invoke vim, can't onvert to WaitStatus, no vim digraphs\n")
				return VimData{}
			}
		default:
			fmt.Fprintf(os.Stderr, "error getting digraphs from vim: %s\n", err)
			return VimData{}
		}
	}

	approxDigraphCountEstimate := 1000

	// Sample: Eu €  8364
	// value is decimal
	digraphExtractor := regexp.MustCompile(`(\S\S)\s+(\S+)\s+(\d+)\b`)
	broken := 0
	results := VimData{
		DigraphByRune: make(map[rune][]VimDigraph, approxDigraphCountEstimate),
	}

	for b.Len() > 0 {
		line, err := b.ReadBytes('\n')
		if err != nil {
			switch err {
			case io.EOF:
				break
			default:
				// FIXME log
				fmt.Fprintf(os.Stderr, "error scanning digraph data which had been retrieved from vim: %s\n", err)
				return VimData{}
			}
		}
		line = line[:len(line)-1]
		segments := bytes.FieldsFunc(line, func(r rune) bool { return r == '\r' })
		for _, segment := range segments {
			if len(segment) < 6 || segment[0] == '\x1B' { // 1B is ESC
				continue
			}
			for _, chunk := range digraphExtractor.FindAllSubmatch(segment, -1) {
				i, err := strconv.Atoi(string(chunk[3]))
				if err != nil {
					broken++
					continue
				}
				theRune, _ := utf8.DecodeRune(chunk[2])
				vd := VimDigraph{
					Sequence:  string(chunk[1]),
					Result:    theRune,
					Codepoint: i,
				}
				// fmt.Fprintf(os.Stderr, "rune: %q gains %v\n", theRune, vd)
				if _, ok := results.DigraphByRune[theRune]; ok {
					results.DigraphByRune[theRune] = append(results.DigraphByRune[theRune], vd)
				} else {
					results.DigraphByRune[theRune] = []VimDigraph{vd}
				}
			}
		}
	}

	// FIXME: log/suppress
	//fmt.Fprintf(os.Stderr,
	//	"scanning digraph data: %d digraphs, %d broken\n",
	//	len(results.DigraphByRune), broken)

	return results
}

// DigraphsFor retrieves a string which is a space-separated list of the known
// digraph sequences which will produce a given rune.
func (v VimData) DigraphsFor(r rune) string {
	items, ok := v.DigraphByRune[r]
	if !ok {
		return ""
	}
	entries := make([]string, 0, len(items))
	uniq := make(map[string]struct{}, len(items))

	for _, digraph := range items {
		if _, ok := uniq[digraph.Sequence]; ok {
			continue
		}
		entries = append(entries, digraph.Sequence)
		uniq[digraph.Sequence] = struct{}{}
	}
	sort.Strings(entries)
	return strings.Join(entries, " ")
}
