// Copyright © 2017,2018 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	// hrm, depending upon a package within which we're generating files
	"github.com/philpennock/character/internal/runemanip"
	"github.com/philpennock/character/unicode"
)

var flags struct {
	outDir                 string
	unstable               bool
	minBlocksFetchSize     uint64
	minUnicodeFetchSize    uint64
	minVariationsFetchSize uint64
	noFetch                bool
	packageName            string
	noFmt                  bool
}
var warningCount int

const (
	stableUnicodeBaseURL   = "https://www.unicode.org/Public/UCD/latest/ucd/"
	unstableUnicodeBaseURL = "https://www.unicode.org/Public/15.0.0/ucd/"
	// FIXME: from Unicode 13.0 onwards, we have "emoji-sequences.txt" instead, in a different format.
	emojiVariationsURL      = "https://unicode.org/Public/emoji/12.1/emoji-variation-sequences.txt"
	blocksFilename          = "Blocks.txt"
	unstableBlocksFilename  = "Blocks-15.0.0.txt"
	blocksOutFilename       = "generated_blocks.go"
	unidataFilename         = "UnicodeData.txt"
	unstableUnidataFilename = "UnicodeData-15.0.0.txt"
	unidataOutFilename      = "generated_data.go"
	emojiVariationsFilename = "emoji-variation-sequences.txt"
	emojiOutFilename        = "generated_emoji.go"

	approxBlockUpperBound      = 500
	approxUnidataUpperBound    = 35000
	approxVariationsUpperBound = 500 // 351 as of emoji 5.0
)

func init() {
	flag.StringVar(&flags.outDir, "output-dir", "unicode", "directory to create files in")
	flag.StringVar(&flags.packageName, "package", "unicode", "package to name generated files")
	flag.BoolVar(&flags.unstable, "unstable", false, "use latest draft Unicode we know of")
	flag.Uint64Var(&flags.minBlocksFetchSize, "min-blocks-fetchsize", 6*1024, "minimum size of Blocks.txt to not be an error")
	flag.Uint64Var(&flags.minUnicodeFetchSize, "min-unicode-fetchsize", 1024*1024, "minimum size of UnicodeData.txt to not be an error")
	flag.Uint64Var(&flags.minVariationsFetchSize, "min-variations-fetchsize", 32*1024, "minimum sie of emoji-variation-sequences.txt to not be an error")
	flag.BoolVar(&flags.noFetch, "no-fetch", false, "do not retrieve current files, regenerate from local only")
	flag.BoolVar(&flags.noFmt, "no-fmt", false, "do not run go fmt automatically")
}

func main() {
	flag.Parse()
	if v := os.Getenv("USE_UNPUBLISHED_UNICODE"); v != "" {
		flags.unstable = true
	}

	if !haveDir(flags.outDir) {
		Die("missing directory %q", flags.outDir)
	}

	var blocksURL, unicodeURL string
	switch flags.unstable {
	case false:
		blocksURL = stableUnicodeBaseURL + blocksFilename
		unicodeURL = stableUnicodeBaseURL + unidataFilename
	case true:
		blocksURL = unstableUnicodeBaseURL + unstableBlocksFilename
		unicodeURL = unstableUnicodeBaseURL + unstableUnidataFilename
	}
	blocksRawOutPath := filepath.Join(flags.outDir, blocksFilename)
	blocksGenOutPath := filepath.Join(flags.outDir, blocksOutFilename)
	unidataRawOutPath := filepath.Join(flags.outDir, unidataFilename)
	unidataGenOutPath := filepath.Join(flags.outDir, unidataOutFilename)
	variationsRawOutPath := filepath.Join(flags.outDir, emojiVariationsFilename)
	variationsGenOutPath := filepath.Join(flags.outDir, emojiOutFilename)

	if !flags.noFetch {
		if err := fetchURLtoFile(blocksURL, blocksRawOutPath, flags.minBlocksFetchSize); err != nil {
			Die("fetching %q from %q failed: %s", blocksRawOutPath, blocksURL, err)
		}
		if err := fetchURLtoFile(unicodeURL, unidataRawOutPath, flags.minUnicodeFetchSize); err != nil {
			Die("fetching %q from %q failed: %s", unidataRawOutPath, unicodeURL, err)
		}
		if err := fetchURLtoFile(emojiVariationsURL, variationsRawOutPath, flags.minVariationsFetchSize); err != nil {
			Die("fetching %q from %q failed: %s", variationsRawOutPath, emojiVariationsURL, err)
		}
	}

	if err := generateBlocksFromTo(blocksRawOutPath, blocksGenOutPath); err != nil {
		Die("Generating %q from %q failed: %s", blocksGenOutPath, blocksRawOutPath, err)
	}

	if err := generateUnicodeDataFromTo(unidataRawOutPath, unidataGenOutPath); err != nil {
		Die("Generating %q from %q failed: %s", unidataGenOutPath, unidataRawOutPath, err)
	}

	if err := generateEmojiVariationsFromTo(variationsRawOutPath, variationsGenOutPath); err != nil {
		Die("Generating %q from %q failed: %s", variationsGenOutPath, variationsRawOutPath, err)
	}

	if !flags.noFmt {
		for _, fn := range []string{blocksGenOutPath, unidataGenOutPath, variationsGenOutPath} {
			if err := reformatFile(fn); err != nil {
				Die("Running go fmt failed: %s", err)
			}
		}
	}

	if warningCount > 0 {
		Warn("encountered %d warnings", warningCount)
		os.Exit(1)
	}
}

func fetchURLtoFile(url, outpath string, minSize uint64) (err error) {
	out, err := os.Create(outpath)
	if err != nil {
		return err
	}
	defer func() {
		e := out.Close()
		if err == nil {
			err = e
		}
	}()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	copied, err := io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	if copied < 0 || uint64(copied) < minSize {
		return fmt.Errorf("copied too little data, only %d octets, need at least %d", copied, minSize)
	}
	Trace("Updated %q from %q, size %d", outpath, url, copied)
	return nil
}

func generateBlocksFromTo(inFn, outFn string) error {
	in, err := os.Open(inFn)
	if err != nil {
		return err
	}
	defer in.Close()

	blocks, maxSeen, err := parseRawBlocks(in)
	if err != nil {
		return err
	}

	out, err := os.Create(outFn)
	if err != nil {
		return err
	}
	defer func() {
		e := out.Close()
		if err == nil {
			err = e
		}
	}()

	nameFilter := strings.NewReplacer(" ", "", "-", "", "_", "")

	fmt.Fprintf(out, "// Code generated by %s; DO NOT EDIT.\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(out, "\npackage %s\n\n", flags.packageName)

	fmt.Fprintf(out, "const (\n")
	for i := range blocks {
		n := "Block" + nameFilter.Replace(blocks[i].Name)
		if i == 0 {
			fmt.Fprintf(out, "\t%s BlockID = iota\n", n)
		} else {
			fmt.Fprintf(out, "\t%s\n", n)
		}
	}
	fmt.Fprintf(out, ")\n")

	fmt.Fprintf(out, "var allKnownBlocks = []BlockInfo{\n")

	for i := range blocks {
		fmt.Fprintf(out, "\tBlockInfo{Min: %d, Max: %d, ID: Block%s, Name: %q},\n",
			blocks[i].Min, blocks[i].Max, nameFilter.Replace(blocks[i].Name), blocks[i].Name)
	}

	fmt.Fprintf(out, "}\n")
	fmt.Fprintf(out, "const maxKnownBlockRune = %d\n", maxSeen)
	return nil
}

func generateUnicodeDataFromTo(inFn, outFn string) error {
	in, err := os.Open(inFn)
	if err != nil {
		return err
	}
	defer in.Close()

	unidata, extra, err := parseUnicodeData(in)
	if err != nil {
		return err
	}

	out, err := os.Create(outFn)
	if err != nil {
		return err
	}
	defer func() {
		e := out.Close()
		if err == nil {
			err = e
		}
	}()

	fmt.Fprintf(out, "// Code generated by %s; DO NOT EDIT.\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(out, "\npackage %s\n\n", flags.packageName)
	fmt.Fprintf(out, "var global = Unicode{\n")
	fmt.Fprintf(out, "// Including ByRune/ByName here hits degenerate compiler edge conditions\n")
	fmt.Fprintf(out, "// such that building this package takes around 40s up from 1s.\n")
	fmt.Fprintf(out, "// So we move that part back to being constructed at runtime.\n\n")

	fmt.Fprintf(out, "\tlinearNames: []string{\n")
	for i := range extra.linearNames {
		fmt.Fprintf(out, "\t\t%q,\n", extra.linearNames[i])
	}
	fmt.Fprintf(out, "\t},\n")
	fmt.Fprintf(out, "\tlinearCI: []CharInfo{\n")
	for i := range extra.linearCI {
		fmt.Fprintf(out, "\t\t%s,\n", fmtCharInfo(extra.linearCI[i]))
	}
	fmt.Fprintf(out, "\t},\n")
	fmt.Fprintf(out, "\tMaxRune: %d,\n", unidata.MaxRune)

	fmt.Fprintf(out, "}\n\n")
	fmt.Fprintf(out, "const runeTotalCount = %d\n", len(extra.linearCI))
	return nil
}

func generateEmojiVariationsFromTo(inFn, outFn string) error {
	in, err := os.Open(inFn)
	if err != nil {
		return err
	}
	defer in.Close()

	emoji, err := parseEmojiData(in)
	if err != nil {
		return err
	}

	out, err := os.Create(outFn)
	if err != nil {
		return err
	}
	defer func() {
		e := out.Close()
		if err == nil {
			err = e
		}
	}()

	fmt.Fprintf(out, "// Code generated by %s; DO NOT EDIT.\n", filepath.Base(os.Args[0]))
	fmt.Fprintf(out, "\npackage %s\n\n", flags.packageName)
	fmt.Fprintf(out, "var emojiable = map[rune]struct{} {\n")
	for _, e := range emoji {
		fmt.Fprintf(out, "\t%d: struct{}{},\n", e)
	}
	fmt.Fprintf(out, "}\n\n")
	fmt.Fprintf(out, "const emojiableTotalCount = %d\n", len(emoji))
	return nil
}

func parseRawBlocks(in io.Reader) ([]unicode.BlockInfo, rune, error) {
	rdr := bufio.NewReader(in)
	ordered := make([]unicode.BlockInfo, 0, approxBlockUpperBound)
	matcher := regexp.MustCompile(`^([0-9A-Fa-f]+)\.\.([0-9A-Fa-f]+);\s+(\S.*?)\s*$`)

	var maxKnownBlockRune rune
	lineNum := 0
ReadLoop:
	for {
		line, err := rdr.ReadBytes('\n')
		lineNum++
		if err != nil {
			switch err {
			case io.EOF:
				break ReadLoop
			default:
				return nil, 0, err
			}
		}
		line = line[:len(line)-1]

		// our embedding inserts an extra newline at the start; be resistant
		if len(line) == 0 {
			continue
		}

		got := matcher.FindSubmatch(line)
		if got == nil {
			continue
		}

		bi := unicode.BlockInfo{
			Min:  runemanip.RuneFromHexField(got[1]),
			Max:  runemanip.RuneFromHexField(got[2]),
			Name: string(got[3]),
		}
		if bi.Max < maxKnownBlockRune {
			return nil, 0, fmt.Errorf("unsorted block info line %d got max %d which < %d", lineNum, bi.Max, maxKnownBlockRune)
		}
		// Trace("Found block %v", bi)
		maxKnownBlockRune = bi.Max
		ordered = append(ordered, bi)
	}

	return ordered, maxKnownBlockRune, nil
}

func fmtCharInfo(ci unicode.CharInfo) string {
	return fmt.Sprintf("CharInfo{Number: %d, Name: %q}", ci.Number, ci.Name)
}

type ExtraUnicode struct {
	// These are for unexported fields in the Unicode
	linearNames []string
	linearCI    []unicode.CharInfo // nb: is `linearIfaceCI []interface{}` in generated code

	// These are just used by us
	Runes []rune
}

func parseUnicodeData(in io.Reader) (unicode.Unicode, ExtraUnicode, error) {
	rdr := bufio.NewReader(in)

	linearNames := make([]string, 0, approxUnidataUpperBound)
	linearCI := make([]unicode.CharInfo, 0, approxUnidataUpperBound)
	Runes := make([]rune, 0, approxUnidataUpperBound)
	var max rune

	lineNum := 0
ReadLoop:
	for {
		line, err := rdr.ReadBytes('\n')
		lineNum++
		if err != nil {
			switch err {
			case io.EOF:
				break ReadLoop
			default:
				return unicode.Unicode{}, ExtraUnicode{}, err
			}
		}
		line = line[:len(line)-1]

		// our embedding inserts an extra newline at the start; be resistant
		if len(line) == 0 {
			continue
		}

		// nb: FieldsFunc collapses sequences of the separator and elides them
		// we want the simpler hard-delimiter split
		fields := bytes.Split(line, []byte{';'})

		r := runemanip.RuneFromHexField(fields[0])
		name := string(fields[1])
		if name == "<control>" {
			if len(fields) >= 11 && len(fields[10]) > 0 {
				name += " [" + string(fields[10]) + "]"
			}
		}
		ci := unicode.CharInfo{
			Number: r,
			Name:   name,
		}
		Runes = append(Runes, r)
		linearNames = append(linearNames, name)
		linearCI = append(linearCI, ci)
		if r > max {
			max = r
		}
	}

	return unicode.Unicode{
			MaxRune: max,
		}, ExtraUnicode{
			linearNames: linearNames,
			linearCI:    linearCI,
			Runes:       Runes,
		}, nil
}

// We return just a list of runes, because as of definition 5.0, the
// definitions consist of two lines per base rune, the first followed by FE0E
// and the second line with the same rune followed by FE0F.  So we will
// complain bitterly if that's not true, because we'll need to update the
// logic, but while that assumption continues to hold true, we just want a
// list.
func parseEmojiData(in io.Reader) ([]rune, error) {
	rdr := bufio.NewReader(in)
	emoji := make([]rune, 0, approxVariationsUpperBound)
	matcher := regexp.MustCompile(`^([0-9A-Fa-f]+)\s+([0-9A-Fa-f]+)\s+;`)

	const (
		textSelect  rune = 0xFE0E
		emojiSelect rune = 0xFE0F
	)

	var textSeen rune
	lineNum := 0
ReadLoop:
	for {
		line, err := rdr.ReadBytes('\n')
		lineNum++
		if err != nil {
			switch err {
			case io.EOF:
				break ReadLoop
			default:
				return nil, err
			}
		}
		line = line[:len(line)-1]

		if len(line) == 0 {
			continue
		}

		got := matcher.FindSubmatch(line)
		if got == nil {
			continue
		}

		baseRune := runemanip.RuneFromHexField(got[1])
		variationSelector := runemanip.RuneFromHexField(got[2])
		switch variationSelector {
		case textSelect:
			if textSeen != 0 {
				return nil, fmt.Errorf("saw two sequential distinct text selectors, %04X then %04X on line %d", textSeen, baseRune, lineNum)
			}
			textSeen = baseRune
		case emojiSelect:
			if textSeen == 0 {
				return nil, fmt.Errorf("saw emoji selector not following a text selector, %04X on line %d", baseRune, lineNum)
			}
			if baseRune != textSeen {
				return nil, fmt.Errorf("saw emoji selector for %04X after non-matching text selector %04X, on line %d", baseRune, textSeen, lineNum)
			}
			emoji = append(emoji, baseRune)
			textSeen = 0
		default:
			return nil, fmt.Errorf("saw unrecognized variation selector %04X on line %d", variationSelector, lineNum)
		}
	}

	if textSeen != 0 {
		return nil, fmt.Errorf("file finished without pairing final rune %04X/text with matching /emoji", textSeen)
	}

	return emoji, nil
}

func reformatFile(fn string) error {
	cmd := exec.Command("go", "fmt", fn)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func haveDir(dirname string) bool {
	st, err := os.Stat(dirname)
	if err != nil {
		if !os.IsNotExist(err) {
			Warn("failed to stat %q: %s", dirname, err)
		}
		return false
	}
	if st.Mode().IsDir() {
		return true
	}
	Warn("not a directory: %q (%s)", dirname, st.Mode())
	return false
}

func Trace(template string, params ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: ", filepath.Base(os.Args[0]))
	fmt.Fprintf(os.Stderr, template, params...)
	fmt.Fprintln(os.Stderr)
}

func Warn(template string, params ...interface{}) {
	Trace(template, params...)
	warningCount++
}

func Die(template string, params ...interface{}) {
	Warn(template, params...)
	os.Exit(1)
}
