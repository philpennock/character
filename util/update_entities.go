// Copyright © 2015 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package main

// +build ignore

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/philpennock/character/aux"
)

func processEntityDirTo(entityFiles []string, outfile string, pkg, mapName string) error {
	entities := make(map[string]rune)
	runes := make(map[rune][]string)

	matcher := regexp.MustCompile(strings.Replace(
		`^\s* \<\!ENTITY \s+ (\S+) \s+ (?:CDATA \s+)? "\&\#(?:[Xx]?)([0-9a-fA-F]+);"`,
		" ", "", -1))

	var lastErr error
	for _, infile := range entityFiles {
		file, err := os.Open(infile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "process: %s\n", err)
			lastErr = err
			continue
		}
		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			got := matcher.FindSubmatch(scanner.Bytes())
			if got == nil {
				continue
			}
			entity := string(got[1])
			r := aux.RuneFromHexField(got[2])
			if have, ok := entities[entity]; ok {
				if have != r {
					fmt.Fprintf(os.Stderr, "duplicate entity definition; %q have %x so ignoring %x\n", entity, have, r)
				}
			} else {
				entities[entity] = r
			}
			if have, ok := runes[r]; ok {
				dup := false
				for _, existing := range have {
					if entity == existing {
						dup = true
						break
					}
				}
				if !dup {
					runes[r] = append(runes[r], entity)
				}
			} else {
				runes[r] = []string{entity}
			}
		}

		if err := scanner.Err(); err != nil {
			fmt.Fprintf(os.Stderr, "scanning %q: %s\n", infile, err)
			if lastErr == nil {
				lastErr = err
			}
		}
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close(%q): %s\n", infile, err)
			if lastErr == nil {
				lastErr = err
			}
		}
	}
	if lastErr != nil {
		return lastErr
	}

	fh, err := os.Create(outfile)
	if err != nil {
		return err
	}

	entityKeys := make([]string, 0, len(entities))
	for k := range entities {
		entityKeys = append(entityKeys, k)
	}
	sort.Strings(entityKeys)
	// will be weirdly sorted for runes with top bit set:
	runeKeys := make([]int, 0, len(runes))
	for k := range runes {
		runeKeys = append(runeKeys, int(k))
	}
	sort.Ints(runeKeys)

	fmt.Fprintf(fh, "package %s\n\nvar %s = map[string]rune{\n", pkg, mapName)
	for _, ent := range entityKeys {
		fmt.Fprintf(fh, "\t\"%s\": %d,\n", ent, entities[ent])
	}
	fmt.Fprintf(fh, "}\n\nvar %sReverse = map[rune][]string{\n", mapName)
	for _, runeI := range runeKeys {
		fmt.Fprintf(fh, "\t%d: []string{\"%s\"},\n", runeI, strings.Join(runes[rune(runeI)], "\", \""))
	}
	fmt.Fprintf(fh, "}\n\n// EOF\n")

	err = fh.Close()
	if err != nil {
		return err
	}

	fmt.Printf("Cleaning up file with go fmt: ")
	cleanup := exec.Command("go", "fmt", outfile)
	cleanup.Stdout = os.Stdout
	cleanup.Stderr = os.Stderr
	return cleanup.Run()
}

var setOfEntities = []struct {
	outfile         string
	mapName         string
	inDirCandidates []string
}{
	{
		"entities/generated_html.go",
		"HtmlEntities",
		[]string{
			"/usr/local/share/sgml/html/4.01",
			"/opt/local/share/OpenSP",       // MacOS(MacPorts)
			"/usr/share/sgml/html/entities", // Debian/Ubuntu
		},
	},
	{
		"entities/generated_xml.go",
		"XmlEntities",
		[]string{
			"/usr/local/share/xml/xmlcharent",
			"/opt/local/share/xml/docbook/*/ent",                 // MacOS(MacPorts)
			"/usr/share/xml/entities/xml-iso-entities-8879.1986", // Debian/Ubuntu
			"/usr/local/Cellar/docbook/*/docbook/xml/*/ent",      // MacOS(homebrew:docbook)
		},
	},
}

// Also: sgml-data gives us /usr/share/sgml/entities/sgml-iso-entities-${versions_list}/ dirs

func findExistingCandidate(list []string) (entityFiles []string) {
	for _, dirOrGlob := range list {
		if strings.ContainsRune(dirOrGlob, '*') {
			dirList, err := filepath.Glob(dirOrGlob)
			if err != nil || dirList == nil {
				continue
			}
			// FIXME: we're assuming version numbers which don't change number
			// of digits (ie that lexicographic sort is also numeric sort).
			// Fix if we ever reach HTML10 or docbook 10.
			sort.Strings(dirList)
			for i := len(dirList) - 1; i >= 0; i-- {
				dir := dirList[i]
				matches, err := filepath.Glob(dir + "/*.ent")
				if err == nil && matches != nil {
					return matches
				}
			}
		} else {
			matches, err := filepath.Glob(dirOrGlob + "/*.ent")
			if err == nil && matches != nil {
				return matches
			}
		}
	}
	return nil
}

func main() {
	for _, s := range setOfEntities {
		entityFiles := findExistingCandidate(s.inDirCandidates)
		if entityFiles == nil {
			fmt.Fprintf(os.Stderr, "unable to find an input dir for %s\n", s.mapName)
			continue
			//os.Exit(1)
		}
		if err := processEntityDirTo(entityFiles, s.outfile, "entities", s.mapName); err != nil {
			fmt.Fprintf(os.Stderr, "making %s: %s\n", s.mapName, err)
			os.Exit(1)
		}
	}
}
