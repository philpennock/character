# Implementation Plan: `character agent` and Additional JSON Fields

This document is the detailed implementation plan for the work described in
`AGENTS.md`.  It supersedes the first-pass plan after two architectural
decisions:

1. **No MCP SDK.** The MCP stdio protocol for a tool-only server is ~200 lines
   of JSON-RPC 2.0 framing.  A hand-rolled `internal/mcpstdio` package keeps
   the project's lean dependency graph intact and keeps the implementation
   fully auditable.

2. **Shared formatting helpers.** Both `resultset.JItem` (CLI `-J` output) and
   `mcpserver.CharProps` (MCP output) compute the same Unicode facts from a
   rune.  Extracting the pure byte-formatting functions into `internal/uformat`
   means the computation lives in one tested place; the two structs remain
   separate because they serve different audiences with different JSON key
   contracts.

---

## Overview

| Part | Summary |
|---|---|
| Pre | `internal/uformat` and `internal/mcpstdio` (no external deps) |
| A | Additional fields in `resultset.JItem` JSON output |
| B | `character agent` sub-command tree (`help`, `examples`, `mcp`) |
| C | MCP stdio server with eight Unicode lookup tools |

**No new external dependencies are required.**

---

## Pre.1 — `internal/uformat`: pure Unicode formatting helpers

**New file: `internal/uformat/uformat.go`**

Pure functions over a single `rune`.  No state, no imports beyond `fmt`,
`strconv`, `strings`, `unicode/utf16`.

- [ ] `func UTF8Bytes(r rune) string`
  Space-separated lowercase two-digit hex bytes.  `0x2713` → `"e2 9c 93"`.
- [ ] `func UTF8Escaped(r rune) string`
  Concatenated `\xNN` per byte, lowercase.  `0x2713` → `"\\xe2\\x9c\\x93"`.
- [ ] `func UnicodeEscaped(r rune) string`
  `r <= 0xFFFF` → `\uXXXX` (four uppercase hex digits).
  `r > 0xFFFF`  → `\UXXXXXXXX` (eight uppercase hex digits).
- [ ] `func RustEscaped(r rune) string`
  `\u{X}` where X is the minimum-length uppercase hex representation of `r`
  (no leading zeros beyond the minimum, e.g. `\u{2713}`, `\u{1F600}`).
- [ ] `func JSONEscaped(r rune) string`
  Mirrors existing `PRINT_RUNE_JSON` logic: UTF-16 surrogate pair if
  necessary; otherwise `\uXXXX`.  Extracts the logic currently embedded in
  `resultset.RenderCharInfoItem` so it can be reused.

**New file: `internal/uformat/uformat_test.go`**

Table-driven tests covering at minimum:

| Rune | Function | Expected |
|---|---|---|
| U+2713 CHECK MARK | `UTF8Bytes` | `"e2 9c 93"` |
| U+2713 | `UTF8Escaped` | `"\\xe2\\x9c\\x93"` |
| U+2713 | `UnicodeEscaped` | `"\\u2713"` |
| U+2713 | `RustEscaped` | `"\\u{2713}"` |
| U+1F600 GRINNING FACE | `UTF8Bytes` | `"f0 9f 98 80"` |
| U+1F600 | `UTF8Escaped` | `"\\xf0\\x9f\\x98\\x80"` |
| U+1F600 | `UnicodeEscaped` | `"\\U0001F600"` |
| U+1F600 | `RustEscaped` | `"\\u{1F600}"` |
| U+0041 LATIN CAPITAL A | `UTF8Bytes` | `"41"` |
| U+0041 | `UnicodeEscaped` | `"\\u0041"` |
| U+1F1FA + U+1F1F8 | `JSONEscaped` | surrogate pair `"\\uD83C\\uDDFA"` |

---

## Pre.2 — `internal/mcpstdio`: hand-rolled MCP stdio server

MCP over stdio is JSON-RPC 2.0 with newline-delimited framing: each message
is a single JSON object on one line terminated by `\n`; messages MUST NOT
contain embedded newlines.  For a tool-only server the required protocol
surface is:

```
← initialize                → InitializeResult (with capabilities.tools)
← notifications/initialized → (no response, ignore)
← tools/list                → ListToolsResult
← tools/call                → CallToolResult
```

**New file: `internal/mcpstdio/mcpstdio.go`**

```go
type Handler func(ctx context.Context, args json.RawMessage) (string, error)

type ToolDef struct {
    Name        string
    Description string
    InputSchema json.RawMessage  // hand-written JSON Schema object
}

type Server struct { ... }

func NewServer(name, version string) *Server
func (s *Server) AddTool(def ToolDef, h Handler)
func (s *Server) ServeStdio(ctx context.Context) error
func (s *Server) ServeConn(ctx context.Context, r io.Reader, w io.Writer) error
```

`ServeStdio` calls `ServeConn(ctx, os.Stdin, os.Stdout)`.
`ServeConn` is the testable entry point.

Internal implementation of `ServeConn`:
- [ ] Framing reader: read one `\n`-terminated line; the line is the JSON body.
- [ ] Framing writer: write the JSON body followed by `\n`.
- [ ] Dispatcher: unmarshal JSON-RPC 2.0 `{"jsonrpc","id","method","params"}`.
  Dispatch on `method`:
  - `"initialize"` → respond with `InitializeResult` (capabilities,
    server info, protocol version `"2024-11-05"`).
  - `"notifications/initialized"` → no response (notification).
  - `"tools/list"` → respond with `{"tools": [...]}` from registered defs.
  - `"tools/call"` → find tool by name, call handler, wrap result.
  - anything else → respond with JSON-RPC error `-32601` (method not found).
- [ ] `tools/call` result wrapping: on handler success, return
  `{"content":[{"type":"text","text":"<result>"}]}`; on handler error, return
  `{"content":[...],"isError":true}`.
- [ ] Request IDs may be numbers, strings, or null; preserve type on echo.
- [ ] Run the read/dispatch/write loop single-threaded (one connection, stdio).
  Use the `context.Context` for cancellation.

**New file: `internal/mcpstdio/mcpstdio_test.go`**

Test using `io.Pipe()` pairs to simulate a client connection without touching
os.Stdin/os.Stdout.  Define a `testClient` helper in the test file:

```go
func testClient(t *testing.T, srv *Server) (send func(method string, params any) json.RawMessage, close func())
```

Pipe-based: `send` writes a framed JSON-RPC request and reads the next framed
response.  `close` shuts down pipes and waits for `ServeConn` to return.

Test cases:

- [ ] Initialize handshake: send `initialize`, assert response contains
  `capabilities.tools` key.
- [ ] `tools/list` after init: assert response `tools` array is non-empty
  when tools have been registered; assert each entry has `name` and
  `inputSchema`.
- [ ] `tools/call` success: register a trivial echo tool; call it; assert
  `content[0].text` equals expected output.
- [ ] `tools/call` handler error: register a tool that returns an error; call
  it; assert `isError == true` in result.
- [ ] `tools/call` unknown tool: assert JSON-RPC error response (not
  `isError`, an actual protocol error).
- [ ] Unknown method: assert JSON-RPC error `-32601`.
- [ ] `notifications/initialized`: assert no response is written (timeout
  after write, then send another request and confirm it is answered).

---

## Part A — Additional JSON fields in `resultset.JItem`

### A.1 — Unicode general category helper

**New file: `unicode/category.go`**

- [ ] Write `func GeneralCategory(r rune) string` returning the two-letter
  Unicode general category abbreviation.
- [ ] Use only the Go standard `unicode` package (no new deps).
- [ ] Iterate an ordered slice of `(abbreviation, *unicode.RangeTable)` pairs —
  Lu, Ll, Lt, Lm, Lo, Mn, Mc, Me, Nd, Nl, No, Pc, Pd, Ps, Pe, Pi, Pf, Po,
  Sm, Sc, Sk, So, Zs, Zl, Zp, Cc, Cf, Cs, Co — calling `unicode.Is` for each.
- [ ] Return `"Cn"` (unassigned) as the fallback when no table matches.
- [ ] Keep the list ordered (do not range over a map) to ensure determinism.

### A.2 — Extend the emoji variation-sequence generator

**Modify: `util/update_unicode.go`**

- [ ] Change the section that reads `emoji-variation-sequences.txt` to build
  two sets: `textable` (U+FE0E entries) and `emojiable` (U+FE0F entries,
  existing).
- [ ] Emit both as `map[rune]struct{}` in the generated output; preserve the
  existing `emojiable` variable name.

**Regenerate: `unicode/generated_emoji.go`**

- [ ] Run `go generate ./unicode/` after changing the generator; commit the
  updated generated file alongside the generator change.

**Extend: `unicode/emoji.go`**

- [ ] Add type `PresentVariant struct { Selector rune; Type string }`.
- [ ] Add `func PresentationVariants(r rune) []PresentVariant` returning nil
  if `r` participates in neither variation, otherwise the applicable entries.

### A.3 — `Blocks.LookupInfo` method

**Modify: `unicode/blocks.go`**

- [ ] Add `func (b Blocks) LookupInfo(r rune) *BlockInfo` returning a pointer
  to a copy of the matching `BlockInfo`, or `nil`.  The existing `Lookup`
  method may delegate to this.

### A.4 — Extend `resultset.JItem` and `JSONEntry`

**Modify: `resultset/resultset.go`**

- [ ] Add new types:

  ```go
  type JBlock struct {
      Name  string `json:"name"`
      Start string `json:"start"` // "U+2700"
      End   string `json:"end"`   // "U+27BF"
  }

  type JPresentVariant struct {
      Selector string `json:"selector"` // "U+FE0F"
      Type     string `json:"type"`     // "text" or "emoji"
  }
  ```

- [ ] Add fields to `JItem` (all `omitempty`; `"block_info"` key preserves the
  existing `"block"` string field without a breaking change):

  ```go
  UTF8Bytes       string            `json:"utf8_bytes,omitempty"`
  UTF8Escaped     string            `json:"utf8_escaped,omitempty"`
  UnicodeEscaped  string            `json:"unicode_escaped,omitempty"`
  RustEscaped     string            `json:"rust_escaped,omitempty"`
  Category        string            `json:"category,omitempty"`
  BlockInfo       *JBlock           `json:"block_info,omitempty"`
  PresentVariants []JPresentVariant `json:"presentation_variants,omitempty"`
  ```

- [ ] Populate new fields in `JSONEntry` using `internal/uformat` functions
  and the new helpers from A.1–A.3.  Skip all new fields when
  `ci.unicode.Number == 0` (synthetic entries for pair/sequence items).
- [ ] Remove the inline `utf16` JSON escape logic from `RenderCharInfoItem`
  for `PRINT_RUNE_JSON` and delegate to `uformat.JSONEscaped` so the
  implementation lives in one place.  The `JSONEscape` field in `JItem` was
  already present; this is a refactor, not a new field.

### A — Tests

**New file: `unicode/category_test.go`**

Table-driven test of `GeneralCategory` (see table in Pre.1 preamble for
representative cases).  Minimum rows: U+0041 → Lu, U+0061 → Ll, U+0030 → Nd,
U+2713 → So, U+0020 → Zs, U+0000 → Cc, U+002E → Po, U+0024 → Sc.

**New file: `unicode/presentation_test.go`**

- [ ] `PresentationVariants(0x2713)` returns one entry, `Type == "emoji"`,
  `Selector == 0xFE0F`.
- [ ] `PresentationVariants(0x0041)` returns nil.
- [ ] `PresentationVariants(0x0023)` (NUMBER SIGN, has both text and emoji
  variants in the Unicode data) returns two entries.

**New or extended `unicode/blocks_info_test.go`**

- [ ] `LookupInfo(0x2713)` returns non-nil; `Name == "Dingbats"`.
- [ ] `LookupInfo(0x0041)` returns non-nil.
- [ ] `LookupInfo(0xFFFF)` does not panic.

**New file: `resultset/json_fields_test.go`**

Uses `sources.NewFast()`.

- [ ] U+2713 CHECK MARK: assert `UTF8Bytes == "e2 9c 93"`,
  `UTF8Escaped == "\\xe2\\x9c\\x93"`, `UnicodeEscaped == "\\u2713"`,
  `RustEscaped == "\\u{2713}"`, `Category == "So"`,
  `BlockInfo.Name == "Dingbats"`, `BlockInfo.Start == "U+2700"`.
- [ ] U+1F600 GRINNING FACE: `UnicodeEscaped == "\\U0001F600"`,
  `RustEscaped == "\\u{1F600}"`, `UTF8Bytes == "f0 9f 98 80"`.
- [ ] U+2603 SNOWMAN: `PresentVariants` non-empty, at least one entry with
  `Selector == "U+FE0F"` and `Type == "emoji"`.
- [ ] U+0041: `PresentVariants` nil/empty.
- [ ] Synthetic entry (Number == 0, e.g. a regional pair result): new fields
  are all empty strings / nil.

---

## Part B — `character agent` sub-command tree

### B.1 — Parent command

**New package: `commands/agent/`**

**New file: `commands/agent/agent.go`**

- [ ] `agentCmd` as `*cobra.Command`, `Use: "agent"`, no `Run`.
- [ ] `root.AddCommand(agentCmd)` in `init()`.

**Modify: `main.go`**

- [ ] Add blank import `_ "github.com/philpennock/character/commands/agent"`.

### B.2 — `agent help` sub-command

**New file: `commands/agent/agenthelp.go`**

- [ ] Output types:

  ```go
  type AgentFlag struct {
      Name        string `json:"name"`
      Short       string `json:"short,omitempty"`
      Type        string `json:"type"`
      Default     string `json:"default"`
      Description string `json:"description"`
  }
  type AgentCommand struct {
      Name        string         `json:"name"`
      Usage       string         `json:"usage"`
      Short       string         `json:"short"`
      Flags       []AgentFlag    `json:"flags,omitempty"`
      SubCommands []AgentCommand `json:"subcommands,omitempty"`
  }
  type AgentHelp struct {
      Tool        string         `json:"tool"`
      Version     string         `json:"version"`
      Description string         `json:"description"`
      Commands    []AgentCommand `json:"commands"`
  }
  ```

- [ ] `Run` walks `root.Cobra().Commands()` recursively.  For each command,
  call `cmd.Flags().VisitAll`; skip hidden flags (`flag.Hidden == true`).
- [ ] Exclude the `agent` command and its children from output.
- [ ] Verify in tests that `--json` appears for `named` (it is a persistent
  flag via `resultset.RegisterCmdFlags`); if `cmd.Flags()` misses it, merge
  `cmd.InheritedFlags()` explicitly.
- [ ] Emit `AgentHelp` as `json.MarshalIndent` to stdout.

### B.3 — `agent examples` sub-command

**New file: `commands/agent/agentexamples.go`**

- [ ] Output type:

  ```go
  type AgentExample struct {
      Category    string `json:"category"`
      Description string `json:"description"`
      Command     string `json:"command"`
      OutputShape string `json:"output_shape"`
  }
  ```

- [ ] Static `[]AgentExample` covering all six categories: `lookup`, `search`,
  `emoji`, `encoding`, `transform`, `browse`.  Minimum two examples per
  category.
- [ ] Optional positional arg `[category]` filters by exact match.  Unknown
  category → empty array, no error.
- [ ] Emit as `json.MarshalIndent`.

### B.4 — `agent mcp` sub-command

**New file: `commands/agent/agentmcp.go`**

- [ ] `Run`:
  1. `sources.NewFast()` then `srcs.LoadUnicodeSearch()` (eager, paid once).
  2. `mcpserver.NewServer(srcs)` (from `commands/agent/mcpserver/`).
  3. `srv.ServeStdio(context.Background())`.
  4. Exit non-zero on error.

### B — Tests

**New file: `commands/agent/agenthelp_test.go`**

- [ ] Capture stdout; invoke `agent help`; unmarshal JSON.
- [ ] Assert `Tool == "character"`, `Commands` non-empty.
- [ ] Find `named` entry; assert `Flags` contains entry with `Name == "json"`,
  `Type == "bool"`.
- [ ] Assert no command named `"agent"` at top level.

**New file: `commands/agent/agentexamples_test.go`**

- [ ] No arg: all six categories present.
- [ ] `"lookup"`: all entries have `Category == "lookup"`.
- [ ] `"nonexistent"`: output is `[]`.

---

## Part C — MCP stdio server

### C.1 — New package `commands/agent/mcpserver`

**New file: `commands/agent/mcpserver/server.go`**

- [ ] `type Server struct` wrapping `*mcpstdio.Server` and `*sources.Sources`.
- [ ] `func NewServer(srcs *sources.Sources) *Server` creates the
  `mcpstdio.Server`, calls `registerTools(srv, srcs)`, returns wrapper.
- [ ] `func (s *Server) ServeStdio(ctx context.Context) error` delegates to
  the embedded `mcpstdio.Server`.
- [ ] `func (s *Server) Inner() *mcpstdio.Server` for test access (so tests
  can call `ServeConn` directly).

### C.2 — Character property type

**New file: `commands/agent/mcpserver/charprops.go`**

- [ ] Define output types:

  ```go
  type CharProps struct {
      Character       string       `json:"character"`
      Name            string       `json:"name"`
      Hex             string       `json:"hex"`
      Decimal         int          `json:"decimal"`
      UTF8Percent     string       `json:"utf8_percent"`
      UTF8Bytes       string       `json:"utf8_bytes"`
      UTF8Escaped     string       `json:"utf8_escaped"`
      UnicodeEscaped  string       `json:"unicode_escaped"`
      RustEscaped     string       `json:"rust_escaped"`
      JSONEscaped     string       `json:"json_escaped"`
      Block           BlockObj     `json:"block"`
      Category        string       `json:"category"`
      RenderWidth     int          `json:"render_width"`
      HTMLEntities    []string     `json:"html_entities,omitempty"`
      XMLEntities     []string     `json:"xml_entities,omitempty"`
      VimDigraphs     []string     `json:"vim_digraphs,omitempty"`
      X11Digraphs     []string     `json:"x11_digraphs,omitempty"`
      PresentVariants []PresentVar `json:"presentation_variants,omitempty"`
  }

  type BlockObj struct {
      Name  string `json:"name"`
      Start string `json:"start"` // "U+2700"
      End   string `json:"end"`   // "U+27BF"
  }

  type PresentVar struct {
      Selector string `json:"selector"` // "U+FE0F"
      Type     string `json:"type"`     // "text" or "emoji"
  }
  ```

- [ ] `func CharPropsFromRune(r rune, srcs *sources.Sources) CharProps`
  computes all fields.  Uses `uformat.*` for byte-formatting fields, the
  unicode package helpers for category and variation, and the sources for
  digraphs and entities.  Must not read or write `resultset.ResultCmdFlags`.

  Field-by-field:
  - `Character`: `string(r)`.
  - `Name`: `srcs.Unicode.ByRune[r].Name`.
  - `Hex`: uppercase `strconv.FormatUint(uint64(r), 16)`.
  - `Decimal`: `int(r)`.
  - `UTF8Percent`: URL-percent encoding (`%E2%9C%93`).
  - `UTF8Bytes`: `uformat.UTF8Bytes(r)`.
  - `UTF8Escaped`: `uformat.UTF8Escaped(r)`.
  - `UnicodeEscaped`: `uformat.UnicodeEscaped(r)`.
  - `RustEscaped`: `uformat.RustEscaped(r)`.
  - `JSONEscaped`: `uformat.JSONEscaped(r)`.
  - `Block`: from `srcs.UBlocks.LookupInfo(r)`; zero `BlockObj` if nil.
  - `Category`: `unicode.GeneralCategory(r)`.
  - `RenderWidth`: `runemanip.DisplayCellWidth(string(r))` first return.
  - `HTMLEntities`, `XMLEntities`: from `entities.*Reverse` maps.
  - `VimDigraphs`, `X11Digraphs`: `srcs.Vim.DigraphsSliceFor(r)` etc.
  - `PresentVariants`: `unicode.PresentationVariants(r)`, selectors as
    `fmt.Sprintf("U+%04X", sel.Selector)`.

### C.3 — Tool input schemas

Each tool's `InputSchema` is a hand-written `json.RawMessage` constant in
`commands/agent/mcpserver/schemas.go`.  Keep them minimal: only list
properties that are actually used.  Example:

```go
var schemaLookupChar = json.RawMessage(`{
  "type": "object",
  "properties": {
    "char": {"type": "string", "description": "A single Unicode character"}
  },
  "required": ["char"],
  "additionalProperties": false
}`)
```

Define one variable per tool.

### C.4 — Tool implementations

**New file: `commands/agent/mcpserver/tools.go`**

`func registerTools(srv *mcpstdio.Server, srcs *sources.Sources)` called by
`NewServer`.  Each handler is a closure over `srcs`.

Handlers receive `json.RawMessage` args, unmarshal into a local struct, do
work, marshal output to JSON, return the JSON string.  Define a helper:

```go
func jsonResult(v any) (string, error) {
    b, err := json.Marshal(v)
    return string(b), err
}
```

#### `unicode_lookup_char`

- [ ] Unmarshal `{"char": "..."}`.
- [ ] Validate: `utf8.RuneCountInString(char) == 1`; error otherwise.
- [ ] Return `jsonResult(CharPropsFromRune(firstRune, srcs))`.

#### `unicode_lookup_name`

- [ ] Unmarshal `{"name": "...", "exact": bool}`.
- [ ] Exact: look up `strings.ToUpper(name)` in `srcs.Unicode.ByName`.
  Return single-element slice or error.
- [ ] Non-exact: `srcs.Unicode.Search.Query(name, -1)`, sort by codepoint,
  return `[]CharProps`.

#### `unicode_search`

- [ ] Unmarshal `{"query": "..."}`.
- [ ] `srcs.Unicode.Search.Query(query, -1)`, sort, return `[]CharProps`.

#### `unicode_lookup_codepoint`

- [ ] Unmarshal `{"codepoint": "..."}`.
- [ ] Parse: strip leading `U+` or `u+` → parse as hex; strip `0x`/`0X` →
  parse as hex; otherwise parse as decimal integer.
- [ ] Return `jsonResult(CharPropsFromRune(rune(n), srcs))`.

#### `unicode_browse_block`

- [ ] Unmarshal `{"block": "..."}`.
- [ ] `srcs.UBlocks.FindByName(block)`: if no exact match, include candidate
  names in error message.
- [ ] Iterate `min..max`; include only runes present in `srcs.Unicode.ByRune`.
- [ ] Hard limit 3000 entries; return error if exceeded.
- [ ] Return `jsonResult([]CharProps{...})`.

#### `unicode_list_blocks`

- [ ] No input schema properties required.
- [ ] `srcs.UBlocks.ListBlocks()`, convert each to `BlockObj`.
- [ ] Return `jsonResult([]BlockObj{...})`.

#### `unicode_emoji_flag`

- [ ] Define result type inline or as a named type:

  ```go
  type FlagResult struct {
      Indicator1 CharProps `json:"indicator_1"`
      Indicator2 CharProps `json:"indicator_2"`
      Combined   string    `json:"combined"`
  }
  ```

- [ ] Unmarshal `{"country_code": "..."}`.
- [ ] Validate exactly two ASCII letters; uppercase.
- [ ] Map each letter to `0x1F1E6 + (letter - 'A')`.
- [ ] Return `FlagResult` with both `CharProps` and `Combined = string(r1)+string(r2)`.

#### `unicode_transform`

**Prerequisite** — export transform functions before implementing this tool:

- [ ] **`commands/transform/fraktur.go`**: add
  `func TransformFraktur(args []string) (string, error)` as a named exported
  wrapper for the existing anonymous Transformer closure.
- [ ] **`commands/transform/turn.go`**: rename `transformTurn` →
  `TransformTurn` (it is already a named function; just capitalise).
- [ ] **`commands/transform/math.go`**: add
  `func TransformMath(args []string, target string) (string, error)` wrapping
  the anonymous Transformer; `target` defaults to regular (empty string) if
  `""` is passed.
- [ ] `scream.NewEncoder()` / `NewDecoder()` are already exported; no change.

Tool handler:

- [ ] Unmarshal `{"type": "...", "text": "...", "target": "..."}`.
- [ ] Dispatch:

  | `type` value | Call |
  |---|---|
  | `"fraktur"` | `transform.TransformFraktur([]string{text})` |
  | `"math"` | `transform.TransformMath([]string{text}, target)` |
  | `"scream"` | `transform.NewEncoder().Replace(text)` |
  | `"scream-decode"` | `transform.NewDecoder().Replace(text)` |
  | `"turn"` | `transform.TransformTurn([]string{text})` |
  | other | return error |

- [ ] Return:

  ```go
  type TransformResult struct {
      Input  string `json:"input"`
      Type   string `json:"type"`
      Output string `json:"output"`
  }
  ```

### C — Tests

**New file: `commands/agent/mcpserver/charprops_test.go`**

Uses `sources.NewFast()`.

- [ ] U+2713: `Name == "CHECK MARK"`, `Hex == "2713"`, `Decimal == 10003`,
  `UTF8Bytes == "e2 9c 93"`, `UnicodeEscaped == "\\u2713"`,
  `RustEscaped == "\\u{2713}"`, `Block.Name == "Dingbats"`,
  `Block.Start == "U+2700"`, `Category == "So"`, `RenderWidth == 1`.
- [ ] U+0041: `Category == "Lu"`, `Name == "LATIN CAPITAL LETTER A"`.
- [ ] U+1F600: `UnicodeEscaped == "\\U0001F600"`, `RustEscaped == "\\u{1F600}"`,
  `UTF8Bytes == "f0 9f 98 80"`.
- [ ] U+1F1EB: `Block.Name` contains `"Regional"`.

**New file: `commands/agent/mcpserver/tools_test.go`**

Test handlers directly: construct `srcs`, call `registerTools` against an
`mcpstdio.Server`, retrieve each handler by iterating registered tools, call
handler with marshaled JSON input, unmarshal the returned string.

Alternatively: use `mcpstdio.Server.ServeConn` with `io.Pipe()` pairs (a
minimal JSON-RPC test client in the test file) for integration-level tests.
Do both: unit tests on handlers, one integration test through `ServeConn`.

Handler unit tests:

- [ ] `unicode_lookup_char("✓")` → `name == "CHECK MARK"`.
- [ ] `unicode_lookup_char("")` → error returned.
- [ ] `unicode_lookup_char("ab")` (two runes) → error returned.
- [ ] `unicode_lookup_codepoint("U+2713")` → `name == "CHECK MARK"`.
- [ ] `unicode_lookup_codepoint("10003")` (decimal) → same.
- [ ] `unicode_lookup_codepoint("0x2713")` → same.
- [ ] `unicode_lookup_name(name="CHECK MARK", exact=true)` → single result.
- [ ] `unicode_lookup_name(name="NONEXISTENT XYZ", exact=true)` → error.
- [ ] `unicode_search("snowman")` → non-empty; contains entry where
  `name == "SNOWMAN"`.
- [ ] `unicode_list_blocks` → non-empty; every entry has non-empty `name`,
  `start`, `end`; `start` matches `^U\+[0-9A-F]+$`.
- [ ] `unicode_browse_block("Dingbats")` → non-empty.
- [ ] `unicode_browse_block("Nonexistent Block XYZ")` → error.
- [ ] `unicode_emoji_flag("FR")` → `combined` non-empty.
- [ ] `unicode_emoji_flag("fr")` (lowercase) → same result.
- [ ] `unicode_emoji_flag("ZZZ")` (three letters) → error.
- [ ] `unicode_transform(type="fraktur", text="Hello")` → `output != input`.
- [ ] `unicode_transform(type="scream")` then
  `unicode_transform(type="scream-decode")` on that output → roundtrip gives
  original input.
- [ ] `unicode_transform(type="invalid")` → error.

`ServeConn` integration test (in `internal/mcpstdio/mcpstdio_test.go`,
covered in Pre.2 above): the protocol-level tests live there, not here.

---

## Implementation order

```
Step 0a  internal/uformat/uformat.go            (pure, no deps)
Step 0b  internal/mcpstdio/mcpstdio.go          (pure, no deps beyond stdlib)
Step 1   unicode/category.go                    (stdlib only)
Step 2   util/update_unicode.go + regenerate    (modify existing generator)
Step 3   unicode/emoji.go: PresentationVariants (needs step 2 output)
Step 4   unicode/blocks.go: LookupInfo          (standalone)
Step 5   resultset/resultset.go: JItem fields   (needs 0a, 1, 3, 4)

Step 6   commands/transform: export API         (standalone refactor)

Step 7   commands/agent/agent.go                (parent, needs root only)
Step 8   commands/agent/agenthelp.go            (needs step 7)
Step 9   commands/agent/agentexamples.go        (needs step 7)

Step 10  commands/agent/mcpserver/charprops.go  (needs 0a, 1, 3, 4)
Step 11  commands/agent/mcpserver/schemas.go    (standalone JSON constants)
Step 12  commands/agent/mcpserver/tools.go      (needs 6, 10, 11)
Step 13  commands/agent/mcpserver/server.go     (needs 0b, 12)
Step 14  commands/agent/agentmcp.go             (needs 7, 13)

Step 15  main.go: blank import                  (needs 7)

Step 16  Tests for steps 0a, 0b, 1–5           (can be written with each step)
Step 17  Tests for steps 7–9
Step 18  Tests for steps 10–14
```

Steps 0a, 0b, 1, 4, 6 are fully independent and can proceed in parallel.
Steps 2 and 3 are sequential.
Steps 8 and 9 are independent once step 7 is done.
Steps 10 and 11 are independent once the Pre and A steps are done.

---

## Architectural notes and risks

### `resultset.ResultCmdFlags` is global state

`ResultSet.New` reads `ResultCmdFlags` globals set by Cobra at parse time.
`CharPropsFromRune` must not use `ResultSet` or `JSONEntry` — those paths are
entangled with CLI flag state that is undefined in the MCP context.  The
separation is enforced by package boundaries: `mcpserver` does not import
`resultset`.

### `JItem` backward compatibility

All new fields use `omitempty`.  The new block object uses key `"block_info"`,
coexisting with the existing `"block"` string field.  The MCP `CharProps`
struct uses `"block"` for the structured object, but it is in a different
package with a different serialisation contract, so there is no collision.

### `uformat` as the single source of truth for byte-formatting

`resultset.JSONEntry` currently has the `utf16`-based JSON escape logic inline.
Step 5 delegates this to `uformat.JSONEscaped`; the inline code is removed.
This is the only instance where Part A changes existing *behaviour* rather than
only adding new fields.  The logic is identical; only the location changes.
Covered by the existing `resultset_test.go` plus the new `json_fields_test.go`.

### MCP search startup cost

`srcs.LoadUnicodeSearch()` (Ferret inverted-suffix index) takes ~100–300 ms.
Called eagerly in `agentmcp.go Run` before `ServeStdio`, so the cost is at
server startup rather than on the first search request.  Document in
`agent mcp --help`.

### Transform package export

`TransformFraktur`, `TransformTurn`, `TransformMath` are currently unexported
closures or private named functions.  Step 6 extracts them to exported named
functions with no behaviour change; the existing Cobra wiring continues to
delegate to the same underlying logic via the same `transformer` type.

### `unicode_browse_block` and CJK gap

CJK Unified Ideographs (U+4E00–U+9FFF) are absent from `srcs.Unicode.ByRune`.
`unicode_browse_block("CJK Unified Ideographs")` returns zero results.  This
is a known limitation matching the CLI.  The tool response should include a
`"note"` field (not a protocol error) when the queried block is within a known
CJK range but yields no results.  This detail may be deferred to a follow-on.

### JSON-RPC `id` field type

JSON-RPC 2.0 allows `id` to be a string, number, or null.  The framing reader
should preserve the raw `id` value (`json.RawMessage`) and echo it unchanged
into responses.  Using `interface{}` and then re-marshalling risks converting
a bare integer id to a float in Go's default JSON decoder.  Use
`json.RawMessage` for the id field throughout `internal/mcpstdio`.
