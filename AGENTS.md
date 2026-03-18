# AGENTS.md — Guide for AI Coding Agents

This file describes the `character` CLI tool for the benefit of AI coding
assistants, LLM-based agents, and automated tooling.  The tool answers
questions about Unicode codepoints, character properties, emoji, and related
transformations.

For understanding the **source code** — package structure, data flow, reading
order, protocols, and architectural invariants — see
[`CODE_GUIDE.md`](CODE_GUIDE.md).


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
| `agent help` | JSON schema of all commands and flags (agent-oriented) |
| `agent examples` | Runnable shell examples for common agent use-cases |
| `agent mcp` | Start an MCP server (stdio) exposing Unicode lookups as tools |
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

### Plain output and the one-line contract

For **`named`**, **`code`**, and **`named -/`** (search): plain output is
exactly one line per result — either the character glyph (default for `named`
and `code`) or the character name.  This contract is strict and reliable for
programmatic use.

For **`name`**: the default is one line per input rune, but two categories of
input cause additional lines to be emitted:

- **Variation selectors** (U+FE00–U+FE0F, U+E0100–U+E01EF): the selector
  itself gets a line, and a second extra line shows the base character and
  selector combined.  So `character name ☃︎` (U+2603 + U+FE0E) emits three
  lines: `SNOWMAN`, `VARIATION SELECTOR-15`, then the combined glyph.

- **Regional indicator pairs** (U+1F1E6–U+1F1FF): each indicator gets a line,
  and a third line summarises the completed pair.  So `character name 🇫🇷`
  emits three lines for the two input codepoints.

Use `-J` / `--json` to avoid any line-count ambiguity in all commands.

### Flags that change output format

| Flag | Effect |
|---|---|
| `-v` | Verbose: render a rich table with all default columns |
| `-N` | Net-verbose: adds network-oriented columns (IDN, punycode) |
| `--json` / `-J` | JSON output — **preferred for programmatic/agent use** |
| `-1` / `--oneline` | All characters on a single line (not available on `name`) |
| `-c` / `--clipboard` | Copy output characters to clipboard (interactive use) |

### JSON output

The `-J` / `--json` flag is the most reliable format for agent consumption.
It removes terminal-rendering concerns (box-drawing characters, column widths)
and gives well-typed fields.  Example:

```
$ character named -Jj SNOWMAN
```

```json
{
  "characters": [
    {
      "display": "☃",
      "name": "SNOWMAN",
      "hex": "2603",
      "decimal": "9731",
      "utf8": "%E2%98%83",
      "block": "Miscellaneous Symbols",
      "htmlEntities": ["snowman"],
      "vimDigraphs": ["sn"],
      "jsonEscape": "\\u2603",
      "renderWidth": 1
    }
  ]
}
```

Key fields in JSON output:

| Field | Content |
|---|---|
| `display` | The character glyph |
| `displayText` | Glyph with text presentation selector appended |
| `displayEmoji` | Glyph with emoji presentation selector appended |
| `name` | Official Unicode name |
| `hex` | Codepoint in hexadecimal (no prefix) |
| `decimal` | Codepoint as decimal string |
| `utf8` | UTF-8 bytes in URL-percent encoding (`%XX%YY%ZZ`) |
| `jsonEscape` | Ready-to-use JSON/JavaScript escape (`\uXXXX` or surrogate pair) |
| `block` | Unicode block name |
| `vimDigraphs` | Vim digraph sequences, if any |
| `x11Digraphs` | X11 compose sequences, if any |
| `htmlEntities` | HTML entity names, if any |
| `xmlEntities` | XML entity names, if any |
| `renderWidth` | Terminal display cell width |
| `part-of` | Present when this entry was decomposed from a larger input |

Use `-J` when processing output in a script or passing it to another tool.


## The `agent` sub-command

`character agent` provides sub-commands whose output is optimised for
consumption by AI agents and tooling.  All output is stable, machine-readable,
and free of ANSI escapes or box-drawing characters.

### `character agent help`

Emits a JSON document describing every command, its flags, their types and
defaults, and brief descriptions.  Intended for agents to self-bootstrap
without reading man pages or README files.

```json
{
  "tool": "character",
  "version": "0.10.0",
  "description": "Unicode codepoint lookup and manipulation tool",
  "commands": [
    {
      "name": "named",
      "usage": "named [name of character]",
      "short": "shows character with given name",
      "plain_output": "one glyph per line (strict: one line per input arg)",
      "flags": [
        { "name": "json",    "short": "J", "type": "bool",   "default": false, "description": "JSON output" },
        { "name": "verbose", "short": "v", "type": "bool",   "default": false, "description": "verbose table" },
        { "name": "join",    "short": "j", "type": "bool",   "default": false, "description": "treat all args as one name" },
        { "name": "search",  "short": "/", "type": "bool",   "default": false, "description": "substring search" }
      ]
    }
  ]
}
```

### `character agent examples [category]`

Emits a JSON array of example invocations, optionally filtered by category.
Each example includes a shell command, a description, and the expected output
shape.

```json
[
  {
    "category": "lookup",
    "description": "Look up a character glyph to get its Unicode name",
    "command": "character name -J ✓",
    "output_shape": "json:characters[0].name"
  },
  {
    "category": "lookup",
    "description": "Find a character by its Unicode name",
    "command": "character named -Jj CHECK MARK",
    "output_shape": "json:characters[0].display"
  },
  {
    "category": "search",
    "description": "Search for characters whose names contain a word",
    "command": "character named -J/ snowman",
    "output_shape": "json:characters[]"
  },
  {
    "category": "emoji",
    "description": "Get a country flag emoji by two-letter country code",
    "command": "character region FR",
    "output_shape": "plain:one line containing the flag glyph"
  },
  {
    "category": "encoding",
    "description": "Get UTF-8 bytes and escape sequences for a codepoint",
    "command": "character code -J U+1F600",
    "output_shape": "json:characters[0].utf8 and .jsonEscape"
  }
]
```

Categories: `lookup`, `search`, `emoji`, `encoding`, `transform`, `browse`.

### `character agent mcp`

Starts an MCP (Model Context Protocol) server on stdio, implementing the
JSON-RPC 2.0 protocol.  MCP-aware coding tools (editors, agents) can connect
and call Unicode lookups as structured tool calls without subprocess management.

#### Exposed MCP tools

| Tool | Parameters | Returns |
|---|---|---|
| `unicode_lookup_char` | `char: string` | Full property object for one codepoint |
| `unicode_lookup_name` | `name: string`, `exact: bool` | Array of matching character objects |
| `unicode_search` | `query: string` | Array of name-matched character objects |
| `unicode_lookup_codepoint` | `codepoint: string` (`"U+2713"` or `"10003"`) | Full property object |
| `unicode_browse_block` | `block: string` | Array of characters in the named block |
| `unicode_list_blocks` | _(none)_ | Array of `{name, start, end}` objects |
| `unicode_emoji_flag` | `country_code: string` | Regional indicator pair + combined glyph |
| `unicode_transform` | `type: string`, `text: string` | Transformed string |

#### Character property object (returned by lookup tools)

```json
{
  "character":       "✓",
  "name":            "CHECK MARK",
  "hex":             "2713",
  "decimal":         10003,
  "utf8_percent":    "%E2%9C%93",
  "utf8_bytes":      "e2 9c 93",
  "utf8_escaped":    "\\xe2\\x9c\\x93",
  "unicode_escaped": "\\u2713",
  "rust_escaped":    "\\u{2713}",
  "json_escaped":    "\\u2713",
  "block":           { "name": "Dingbats", "start": "U+2700", "end": "U+27BF" },
  "category":        "So",
  "render_width":    1,
  "html_entities":   ["checkmark", "check"],
  "xml_entities":    [],
  "vim_digraphs":    ["OK"],
  "x11_digraphs":    ["checkmark"],
  "presentation_variants": [
    { "selector": "U+FE0F", "type": "emoji" }
  ]
}
```

The `utf8_escaped`, `unicode_escaped`, `rust_escaped`, and `json_escaped`
fields are ready for direct insertion into source code strings in their
respective languages without further transformation.

## Recommended patterns for agents

### "I have a character glyph; what is it?"

```sh
character name -J ✓
# → JSON: .characters[0].name = "CHECK MARK"
```

### "I want to use a character by name in my code"

```sh
character named -Jj LATIN SMALL LETTER A WITH DIAERESIS
# → JSON: .characters[0].unicode_escaped = "\\u00e4"
# → JSON: .characters[0].utf8_escaped    = "\\xc3\\xa4"
```

### "What characters are called something like 'check'?"

```sh
character named -J/ check
# → JSON: .characters[] each with .name and .display
```

### "What is in the Arrows block?"

```sh
character browse -b Arrows
# → verbose table; for JSON use 'character agent mcp' unicode_browse_block

character known -b
# → list of all block names (use to discover exact spellings)
```

### "I need a flag emoji for France"

```sh
character region FR
# → 🇫🇷  (one line: the combined regional indicator pair)
```

### "Give me the UTF-8 bytes for U+1F600"

```sh
character code -J U+1F600
# → .characters[0].utf8        = "%F0%9F%98%80"
# → .characters[0].utf8_bytes  = "f0 9f 98 80"
# → .characters[0].utf8_escaped = "\\xf0\\x9f\\x98\\x80"
```

### "I need to combine an emoji with a skin-tone modifier"

```sh
character named -1j 'RAISED HAND' 'EMOJI MODIFIER FITZPATRICK TYPE-4'
# → 🤚🏽  (combined on one line, ready to paste)
```


## Important behaviours and limitations

- **CJK Unified Ideographs** (e.g. kanji): the tool returns the block name and
  codepoint but no name, since the Unicode standard does not assign individual
  names to that range.  This is not an error.
- **Substring search** (`named -/` or `search`) uses an inverted-suffix index
  built lazily on first use; the first call may be slower.
- **Variation selectors**: see "One-line contract" above.  For clean parsing,
  use `-J`.
- **Regional indicators**: flag emoji are encoded as pairs of Regional
  Indicator letters.  Use `character region <CC>` for flag lookup; use
  `character name` on existing glyphs to decompose them.
- The CLI JSON output (`-J`) includes `utf8_bytes`, `utf8_escaped`,
  `unicode_escaped`, `rust_escaped`, `category`, `block_info` (object), and
  `presentation_variants`.  The MCP `CharProps` output includes all of these
  plus additional fields; see the property object table above.


## Concise invocation examples for copy-paste

```sh
# Agent-oriented help
character agent help
character agent examples
character agent examples emoji

# Character info — JSON (reliable, parseable)
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
