package unicode

import (
	"bytes"
	"io"
	"sync"
)

// CharInfo is the basic set of information about one Unicode character.
// We record the codepoint (as a Go rune) and the formal Name.
type CharInfo struct {
	Number rune
	Name   string
}

// Unicode is the set of all data about all characters which we've retrieved
// from formal Unicode specifications.
type Unicode struct {
	ByRune map[rune]CharInfo
	ByName map[string]CharInfo

	// should also add an inverted index by word, etc
}

var global Unicode
var parseUnicodeOnce sync.Once

// Load gives us all the Unicode-spec derived data which we have.
func Load() Unicode {
	parseUnicodeOnce.Do(parseRaw)
	return global
}

func parseRaw() {
	b := bytes.NewBuffer(rawData)

	byRune := make(map[rune]CharInfo, rawLineCount)
	byName := make(map[string]CharInfo, rawLineCount)

	lineNum := 0
	for {
		if b.Len() == 0 {
			break
		}
		line, err := b.ReadBytes('\n')
		lineNum++
		if err != nil {
			switch err {
			case io.EOF:
				break
			default:
				panic(err.Error())
			}
		}
		line = line[:len(line)-1]

		// our embedding inserts an extra newline at the start; be resistant
		if len(line) == 0 {
			continue
		}

		fields := bytes.FieldsFunc(line, func(r rune) bool { return r == ';' })

		r := runeFromHexField(fields[0])
		name := string(fields[1])
		ci := CharInfo{
			Number: r,
			Name:   name,
		}
		byRune[r] = ci
		byName[name] = ci
	}

	global = Unicode{
		ByRune: byRune,
		ByName: byName,
	}
}

func runeFromHexField(bb []byte) rune {
	// fields[0] is the hex encoding, but with perhaps odd numbers of bytes present (eg, 5)
	// So rather than `hex.Decode()`, we just decode manually
	var r rune
	for _, c := range bb {
		r *= 16
		switch {
		case '0' <= c && c <= '9':
			r += rune(c - '0')
		case 'A' <= c && c <= 'F':
			r += rune(10 + c - 'A')
		case 'a' <= c && c <= 'f':
			r += rune(10 + c - 'a')
		}
	}
	return r
}
