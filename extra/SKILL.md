# character — Unicode MCP Tool Skill

You have access to a `character` MCP server that provides Unicode character
lookup, search, and transformation tools.  This file teaches you how to use
them correctly.


## Critical: parameter names come from the schema

Every tool has a JSON Schema that defines its **exact** parameter names.
Do not guess parameter names from tool names or descriptions — always use the
names listed below.  For example, `unicode_browse_block` takes a parameter
called `block`, **not** `block_name`.


## Tool reference

### unicode_lookup_char

Look up a single character by its glyph.

```json
{ "char": "✓" }
```

- `char` (string, required): exactly one Unicode codepoint.
- Returns: a single character property object.

### unicode_lookup_name

Look up characters by their Unicode name.

```json
{ "name": "CHECK MARK", "exact": true }
{ "name": "check", "exact": false }
```

- `name` (string, required): the name to look up.
- `exact` (boolean, optional, default false): if true, match the full official
  Unicode name exactly (case-insensitive); if false, perform substring search.
- Returns: array of character property objects.

### unicode_search

Search for characters whose names contain a substring.

```json
{ "query": "snowman" }
```

- `query` (string, required): substring to match against character names.
- Returns: array of character property objects.

### unicode_lookup_codepoint

Look up a character by its codepoint number.

```json
{ "codepoint": "U+2713" }
{ "codepoint": "0x2713" }
{ "codepoint": "10003" }
```

- `codepoint` (string, required): codepoint in `U+XXXX`, `0xXXXX`, or decimal.
- Returns: a single character property object.

### unicode_browse_block

List all characters in a Unicode block.

```json
{ "block": "Dingbats" }
{ "block": "Miscellaneous Symbols" }
```

- `block` (string, required): block name, case-insensitive, partial match
  accepted.  Use `unicode_list_blocks` to discover exact block names.
- Returns: array of character property objects (max 3000).
- Blocks with more than 3000 assigned characters (e.g. CJK Unified
  Ideographs) will return an error — use `unicode_lookup_codepoint` for
  individual characters in large blocks.

### unicode_list_blocks

List all Unicode blocks with their codepoint ranges.  Takes no parameters.

```json
{}
```

- Returns: array of `{ "name", "start", "end" }` objects.

### unicode_emoji_flag

Get a country flag emoji from a two-letter country code.

```json
{ "country_code": "FR" }
```

- `country_code` (string, required): ISO 3166-1 alpha-2 code (case-insensitive).
- Returns: the two regional indicator characters and their combined glyph.

### unicode_transform

Apply a text transformation.

```json
{ "type": "fraktur", "text": "Hello World" }
{ "type": "math", "text": "Hello", "target": "bold" }
{ "type": "scream", "text": "hello world" }
{ "type": "turn", "text": "Hello" }
```

- `type` (string, required): one of `fraktur`, `math`, `scream`,
  `scream-decode`, `turn`.
- `text` (string, required): the text to transform.
- `target` (string, optional): for `math` transforms, the variant name
  (e.g. `bold`, `italic`, `frakturnormal`); defaults to `normal`.
- Returns: `{ "input", "type", "output" }`.


## Character property object

Every lookup tool returns objects with these fields:

| Field | Type | Example | Description |
|---|---|---|---|
| `character` | string | `"✓"` | The character glyph |
| `name` | string | `"CHECK MARK"` | Official Unicode name |
| `hex` | string | `"2713"` | Codepoint in hex (no prefix) |
| `decimal` | int | `10003` | Codepoint as decimal |
| `utf8_percent` | string | `"%E2%9C%93"` | URL percent-encoded UTF-8 bytes |
| `utf8_bytes` | string | `"e2 9c 93"` | Space-separated UTF-8 hex bytes |
| `utf8_escaped` | string | `"\\xe2\\x9c\\x93"` | C-style byte escapes |
| `unicode_escaped` | string | `"\\u2713"` | `\uXXXX` or `\UXXXXXXXX` |
| `rust_escaped` | string | `"\\u{2713}"` | Rust Unicode escape |
| `json_escaped` | string | `"\\u2713"` | JSON `\uXXXX` (surrogate pairs for non-BMP) |
| `block` | object | `{"name":"Dingbats","start":"U+2700","end":"U+27BF"}` | Unicode block |
| `category` | string | `"So"` | General Category abbreviation |
| `render_width` | int | `1` | Terminal display cell width |
| `html_entities` | string[] | `["checkmark","check"]` | HTML entity names (if any) |
| `xml_entities` | string[] | `[]` | XML entity names (if any) |
| `vim_digraphs` | string[] | `["OK"]` | Vim digraph sequences (if any) |
| `x11_digraphs` | string[] | `["checkmark"]` | X11 compose sequences (if any) |
| `presentation_variants` | object[] | `[{"selector":"U+FE0F","type":"emoji"}]` | Variation selectors (if any) |


## Which tool to use

| I want to… | Tool | Example args |
|---|---|---|
| Identify a character I can see | `unicode_lookup_char` | `{"char":"✓"}` |
| Find a character by its exact name | `unicode_lookup_name` | `{"name":"SNOWMAN","exact":true}` |
| Search names by substring | `unicode_search` | `{"query":"arrow"}` |
| Look up a codepoint I already know | `unicode_lookup_codepoint` | `{"codepoint":"U+2603"}` |
| See what characters are in a block | `unicode_browse_block` | `{"block":"Arrows"}` |
| Find out what blocks exist | `unicode_list_blocks` | `{}` |
| Get a country flag emoji | `unicode_emoji_flag` | `{"country_code":"JP"}` |
| Transform text stylistically | `unicode_transform` | `{"type":"fraktur","text":"hi"}` |


## Tips

- **Name search is substring-based**: `{"query":"arrow"}` matches LEFTWARDS
  ARROW, RIGHTWARDS ARROW, etc.  Use `unicode_search` for broad discovery and
  `unicode_lookup_name` with `exact: true` when you know the full name.
- **Block names** are the official Unicode block names.  Call
  `unicode_list_blocks` first if you are unsure of the exact spelling.
- **CJK blocks are large**: `unicode_browse_block` enforces a 3000-character
  limit.  For CJK Unified Ideographs, look up individual codepoints instead.
- **Escape fields are insertion-ready**: `json_escaped`, `unicode_escaped`,
  `rust_escaped`, and `utf8_escaped` can be pasted directly into source code
  strings in their respective languages.
- **Presentation variants**: some characters (e.g. ☃ SNOWMAN) have both text
  and emoji presentation forms.  The `presentation_variants` field lists the
  available variation selectors.
