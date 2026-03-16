# AGENTS.md — Guide for AI Coding Agents

This file describes the `character` CLI tool for the benefit of AI coding
assistants, LLM-based agents, and automated tooling.  The tool answers
questions about Unicode codepoints, character properties, emoji, and related
transformations.


## What this tool does

`character` is a Unicode character-information CLI.  Given a character, a
name, or a codepoint number, it returns authoritative Unicode metadata:
official name, hex codepoint, UTF-8 encoding bytes, block membership, HTML/XML
entity aliases, Vim digraphs, and more.

Common uses for an AI agent:
- Look up the canonical name of a character so you can refer to it precisely.
- Find a character by name or substring when you know what it looks like but
  not its codepoint.
- Verify the exact bytes that will appear in source code when using a Unicode
  literal.
- Look up emoji codepoints and their modifier sequences before writing them
  into code, comments, or user-facing strings.
- Get the right representation (text vs. emoji presentation variant) for a
  character.


## Quick-reference: key sub-commands

| Sub-command | Purpose |
|---|---|
| `name <char>…` | Properties of literal character(s) |
| `named <NAME>` | Look up a character by its exact Unicode name |
| `named -j <MULTI WORD NAME>` | Join all args as one name |
| `named -/ <word>` | Substring search across all character names |
| `search <word>` | Alias for `named -v/` (verbose substring search, table) |
| `code <U+XXXX>` or `code <decimal>` | Look up by codepoint |
| `browse -b <block-name>` | List all characters in a Unicode block |
| `browse -f <U+X> -t <U+Y>` | List characters in a codepoint range |
| `transform fraktur <text>` | Convert text to Fraktur mathematical letters |
| `transform math <text>` | Mathematical letter variants |
| `transform scream <text>` | Reversible "scream" encoding |
| `transform turn <text>` | Upside-down character transformation |
| `region <XY>` | Convert two-letter country code to flag emoji regional indicators |
| `x-puny <string>` | Punycode encode/decode |
| `known -b` | List all Unicode block names |
| `known -e` | List supported character-set encoding names |
| `version -j` | Version info as JSON |


## Output formats

By default most commands print one line per character (the name, or the
character glyph).

### Flags that change output format

| Flag | Effect |
|---|---|
| `-v` | Verbose: render a rich table with all default columns |
| `-N` | Net-verbose: adds network-oriented columns (IDN, punycode) |
| `--json` / `-J` | JSON output — **preferred for programmatic/agent use** |
| `-1` / `--oneline` | All characters on a single line (no separators) |
| `-c` / `--clipboard` | Copy output characters to clipboard (interactive use) |

### JSON output

The `-J` / `--json` flag is the most reliable format for agent consumption.
It removes terminal-rendering concerns (box-drawing characters, ANSI, column
widths) and gives well-typed fields.  Example:

```
$ character named -Jj SNOWMAN
```

```json
{
  "characters": [
    {
      "character": "☃",
      "name": "SNOWMAN",
      "hex": "2603",
      "decimal": 9731,
      "utf8": "e2 98 83",
      ...
    }
  ]
}
```

Use `-J` when processing output in a script or passing it to another tool.


## Recommended patterns for agents

### "I have a character glyph; what is it?"

```sh
character name ☃
# → SNOWMAN  (plain, one line)

character name -J ☃
# → JSON with all properties
```

### "I want to use a character by name in my code"

```sh
character named -j LATIN SMALL LETTER A WITH DIAERESIS
# → ä   (plain output — the literal glyph)

character named -Jj LATIN SMALL LETTER A WITH DIAERESIS
# → JSON
```

### "What characters are called something like 'check'?"

```sh
character search check
# → table of all characters whose names contain "check"

character named -J/ check
# → same results as JSON
```

### "What is in the Arrows block?"

```sh
character browse -b Arrows
# → table of all arrow characters

character known -b
# → list of all block names (use to discover exact block names)
```

### "I need a flag emoji for France"

```sh
character region FR
# → 🇫🇷
```

### "Give me the UTF-8 bytes for U+1F600"

```sh
character code -J U+1F600
# → JSON includes utf8 field with space-separated hex bytes
```

### "I need to combine an emoji with a skin-tone modifier"

```sh
character named -1 'RAISED HAND' 'EMOJI MODIFIER FITZPATRICK TYPE-4'
# → 🤚🏽  (combined on one line, ready to paste)
```


## Important behaviours and limitations

- **CJK Unified Ideographs** (e.g. kanji): the tool returns the block name and
  codepoint but no name, since the Unicode standard does not assign individual
  names to that range.  This is not an error.
- **Substring search** (`named -/` or `search`) uses an inverted-suffix index
  that is built lazily on first use; the first search may be slower.
- **Variation selectors**: U+FE0E forces text presentation, U+FE0F forces emoji
  presentation.  Use `character name` on them to understand what you are
  dealing with.
- **Regional indicators**: flag emoji are encoded as pairs of Regional Indicator
  letters (U+1F1E6–U+1F1FF).  Use `character region <CC>` to get the correct
  pair for a country code.
- **Encoding columns** (visible with `-v`): shows UTF-8 bytes, HTML entity,
  XML entity, and Vim digraph where applicable.


## Concise invocation examples for copy-paste

```sh
# Character info — JSON
character name -J ✓
character named -Jj CHECK MARK
character code -J U+2713

# Search — JSON
character named -J/ checkmark

# Emoji with modifier — one line
character named -1j 'WOMAN SHRUGGING' 'EMOJI MODIFIER FITZPATRICK TYPE-3'

# Browse a block — table
character browse -b 'Miscellaneous Symbols'

# Flag emoji
character region JP

# Transform
character transform fraktur 'Hello World'

# List Unicode blocks
character known -b

# Version
character version -j
```
