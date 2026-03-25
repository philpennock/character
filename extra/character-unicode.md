# character — Unicode MCP Tool Slash Command

Detailed usage guide for the `character` MCP server's Unicode tools.
This supplements the server's built-in `instructions` field (which covers
discovery and key usage patterns) with tool routing, the full property field
reference, and advanced tips.


## Critical: parameter names come from the schema

Every tool has a JSON Schema that defines its **exact** parameter names.
Do not guess parameter names from tool names or descriptions — always use the
names listed in the schema.  For example, `unicode_browse_block` takes a
parameter called `block`, **not** `block_name`.


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


## Character property object

Lookup tools return objects with these fields:

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


## Tips

- **Name search is substring-based**: `{"query":"arrow"}` matches LEFTWARDS
  ARROW, RIGHTWARDS ARROW, etc.  Use `unicode_search` for broad discovery and
  `unicode_lookup_name` with `exact: true` when you know the full name.
- **Escape fields are insertion-ready**: `json_escaped`, `unicode_escaped`,
  `rust_escaped`, and `utf8_escaped` can be pasted directly into source code
  strings in their respective languages.
- **Presentation variants**: some characters (e.g. ☃ SNOWMAN) have both text
  and emoji presentation forms.  The `presentation_variants` field lists the
  available variation selectors.
