# Implementation Plan: `character agent` and Additional JSON Fields

This document is the detailed implementation plan for the work described in
`AGENTS.md`.  It is divided into three parts, with a dependency-ordered
checklist and test specifications for each.

---

## Overview

| Part | Summary |
|---|---|
| A | Additional fields in `resultset.JItem` JSON output |
| B | `character agent` sub-command tree (`help`, `examples`, `mcp`) |
| C | MCP stdio server with eight Unicode lookup tools |

**Prerequisite — new Go dependency** (needed by Part C):

The official Go MCP SDK is already in the local module cache at
`github.com/modelcontextprotocol/go-sdk v0.8.0` and requires no network
access.

```
go get github.com/modelcontextprotocol/go-sdk@v0.8.0
go mod tidy
```

Run this before starting Part C work.

---

## Part A — Additional JSON fields in `resultset.JItem`

### A.1 — Unicode general category helper

**New file: `unicode/category.go`**

- [ ] Write `func GeneralCategory(r rune) string` returning a two-letter
  Unicode general category abbreviation.
- [ ] Use only the Go standard `unicode` package (no new deps).
- [ ] Iterate an ordered slice of `(abbreviation, *unicode.RangeTable)` pairs
  — Lu, Ll, Lt, Lm, Lo, Mn, Mc, Me, Nd, Nl, No, Pc, Pd, Ps, Pe, Pi, Pf, Po,
  Sm, Sc, Sk, So, Zs, Zl, Zp, Cc, Cf, Cs, Co — calling `unicode.Is` for each.
- [ ] Return `"Cn"` (unassigned) as the fallback when no table matches.
- [ ] Keep the list ordered (do not range over a map) to ensure determinism.

### A.2 — Extend the emoji variation-sequence generator

**Modify: `util/update_unicode.go`**

- [ ] Change the section that reads `emoji-variation-sequences.txt` to build
  two separate sets:
  - `textable` — runes with a text-variation entry (selector U+FE0E).
  - `emojiable` — runes with an emoji-variation entry (selector U+FE0F,
    existing behaviour).
- [ ] Emit both as `map[rune]struct{}` variables in the generated output.
  The existing `emojiable` variable name must be preserved; the new variable
  is `textable`.

**Regenerate: `unicode/generated_emoji.go`**

- [ ] Run `go generate ./unicode/` after the above change and commit the
  updated generated file alongside the generator change.

**Extend: `unicode/emoji.go`**

- [ ] Add `func PresentationVariants(r rune) []PresentVariant` (use the type
  defined in A.3 below, or define a local equivalent).  Returns `nil` if `r`
  participates in neither variation.

  ```go
  type PresentVariant struct {
      Selector rune   // U+FE0E or U+FE0F
      Type     string // "text" or "emoji"
  }
  ```

### A.3 — `Blocks.LookupInfo` method

**Modify: `unicode/blocks.go`**

- [ ] Add `func (b Blocks) LookupInfo(r rune) *BlockInfo` returning a pointer
  to a copy of the matching `BlockInfo`, or `nil` if no block covers `r`.
  The existing `Lookup` method can delegate to this.

### A.4 — Extend `resultset.JItem` and `JSONEntry`

**Modify: `resultset/resultset.go`**

- [ ] Add new types alongside the existing ones:

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

- [ ] Add the following fields to `JItem`.  All use `omitempty` to avoid
  breaking existing consumers.  The new block field uses key `"block_info"` to
  coexist with the existing `"block"` string field without a breaking change:

  ```go
  UTF8Bytes       string           `json:"utf8_bytes,omitempty"`
  UTF8Escaped     string           `json:"utf8_escaped,omitempty"`
  UnicodeEscaped  string           `json:"unicode_escaped,omitempty"`
  RustEscaped     string           `json:"rust_escaped,omitempty"`
  Category        string           `json:"category,omitempty"`
  BlockInfo       *JBlock          `json:"block_info,omitempty"`
  PresentVariants []JPresentVariant `json:"presentation_variants,omitempty"`
  ```

- [ ] Populate new fields in `JSONEntry`.  Computations:
  - `UTF8Bytes`: `[]byte(string(r))`; format each byte as two lowercase hex
    digits joined with spaces (e.g. `"e2 9c 93"`).
  - `UTF8Escaped`: same bytes formatted as `\xNN` lowercase hex concatenated
    (e.g. `"\\xe2\\x9c\\x93"`).
  - `UnicodeEscaped`: `r <= 0xFFFF` → `\uXXXX` (four uppercase hex digits);
    otherwise `\UXXXXXXXX` (eight uppercase hex digits).
  - `RustEscaped`: `\u{XXXX}` where XXXX is the minimum-width uppercase hex
    representation of `r` (no leading zeros beyond minimum).
  - `Category`: call `unicode.GeneralCategory(r)`.
  - `BlockInfo`: call `rs.sources.UBlocks.LookupInfo(r)`, and if non-nil
    format `Start` and `End` as `fmt.Sprintf("U+%04X", bi.Min)` etc.
  - `PresentVariants`: call `unicode.PresentationVariants(r)`, convert each
    `PresentVariant.Selector` to `fmt.Sprintf("U+%04X", sel)` for the JSON
    field.
  - Skip all new fields (leave zero value) when `ci.unicode.Number == 0`
    (synthetic entries for combinations/pairs).

- [ ] Note: `UTF8Bytes` for Number==0 (string-sequence entries) should also be
  omitted; the `omitempty` tag handles this automatically for strings.

### A — Tests

**New file: `unicode/category_test.go`**

Table-driven test of `GeneralCategory`:

| Rune | Expected |
|---|---|
| U+0041 LATIN CAPITAL LETTER A | `"Lu"` |
| U+0061 LATIN SMALL LETTER A | `"Ll"` |
| U+0030 DIGIT ZERO | `"Nd"` |
| U+2713 CHECK MARK | `"So"` |
| U+0020 SPACE | `"Zs"` |
| U+0000 NULL | `"Cc"` |
| U+002E FULL STOP | `"Po"` |
| U+0024 DOLLAR SIGN | `"Sc"` |

**New file: `resultset/json_fields_test.go`**

- [ ] Call `sources.NewFast()`, build a `ResultSet`, add U+2713 CHECK MARK.
- [ ] Assert `UTF8Bytes == "e2 9c 93"`.
- [ ] Assert `UTF8Escaped == "\\xe2\\x9c\\x93"`.
- [ ] Assert `UnicodeEscaped == "\\u2713"`.
- [ ] Assert `RustEscaped == "\\u{2713}"`.
- [ ] Assert `Category == "So"`.
- [ ] Assert `BlockInfo != nil`, `BlockInfo.Name == "Dingbats"`,
  `BlockInfo.Start == "U+2700"`, `BlockInfo.End == "U+27BF"`.
- [ ] Repeat for U+1F600 GRINNING FACE:
  - `UnicodeEscaped == "\\U0001F600"` (8-digit form for supplementary plane).
  - `RustEscaped == "\\u{1F600}"`.
  - `UTF8Bytes == "f0 9f 98 80"`.
- [ ] Assert U+2603 SNOWMAN has `PresentVariants` with at least one entry with
  `Selector == "U+FE0F"` and `Type == "emoji"`.
- [ ] Assert U+0041 LATIN CAPITAL LETTER A has nil/empty `PresentVariants`.

**Extend `unicode/emoji_test.go` (or new file `unicode/presentation_test.go`)**

- [ ] `PresentationVariants(0x2713)` returns a non-nil slice with one entry,
  `Type == "emoji"`, `Selector == 0xFE0F`.
- [ ] `PresentationVariants(0x0041)` returns nil.
- [ ] `PresentationVariants(0x0023)` (NUMBER SIGN, has both text and emoji
  variants) returns two entries.

**Extend `unicode/blocks_test.go` (or new `unicode/blocks_info_test.go`)**

- [ ] `LookupInfo(0x2713)` returns non-nil, Name `"Dingbats"`.
- [ ] `LookupInfo(0x0000)` returns non-nil (Basic Latin or Controls).
- [ ] `LookupInfo(0xFFFF)` — either nil or a valid block; assert no panic.

---

## Part B — `character agent` sub-command tree

### B.1 — Parent command

**New package: `commands/agent/`**

**New file: `commands/agent/agent.go`**

- [ ] Define `agentCmd` as a `*cobra.Command` with `Use: "agent"` and no `Run`
  (sub-commands only).
- [ ] Register with `root.AddCommand(agentCmd)` in `init()`.

**Modify: `main.go`**

- [ ] Add blank import `_ "github.com/philpennock/character/commands/agent"` so
  the package's `init()` chain fires.

### B.2 — `agent help` sub-command

**New file: `commands/agent/agenthelp.go`**

- [ ] Define local output types:

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

- [ ] In `Run`: walk `root.Cobra().Commands()` recursively using a helper
  `buildAgentCommand(cmd *cobra.Command) AgentCommand`.
- [ ] In `buildAgentCommand`: call `cmd.Flags().VisitAll` to collect flags.
  Skip hidden flags (`flag.Hidden == true`). Extract type from
  `flag.Value.Type()`. Extract default from `flag.DefValue`.
- [ ] Important: `cmd.Flags()` on a Cobra command includes inherited
  persistent flags after the command is fully initialised. Verify in tests
  that `--json` appears for the `named` sub-command entry (it is registered as
  a persistent flag via `resultset.RegisterCmdFlags`).  If absent, switch to
  merging `cmd.Flags()` and `cmd.InheritedFlags()` explicitly.
- [ ] Skip the `agent` command and its children from the output (they are
  meta-commands, not user-facing Unicode commands, and including them would be
  circular).
- [ ] Emit `AgentHelp` as `json.MarshalIndent` to `os.Stdout`.
- [ ] Register as a sub-command of `agentCmd` in `init()`.

### B.3 — `agent examples` sub-command

**New file: `commands/agent/agentexamples.go`**

- [ ] Define output type:

  ```go
  type AgentExample struct {
      Category    string `json:"category"`
      Description string `json:"description"`
      Command     string `json:"command"`
      OutputShape string `json:"output_shape"`
  }
  ```

- [ ] Define a static `[]AgentExample` slice covering all six categories:
  `lookup`, `search`, `emoji`, `encoding`, `transform`, `browse`.  Draw the
  content from `AGENTS.md` §"Concise invocation examples".  Aim for at least
  two examples per category.
- [ ] Accept an optional positional argument `[category]`.  If present, filter
  the slice to only entries where `Category == arg`.  If the category is
  unknown, return an empty array (not an error).
- [ ] Emit as `json.MarshalIndent` of the (possibly filtered) slice.
- [ ] Register as a sub-command of `agentCmd` in `init()`.

### B.4 — `agent mcp` stub sub-command

**New file: `commands/agent/agentmcp.go`**

(Full MCP implementation is in Part C; this file is the Cobra entry point.)

- [ ] Define a `Run` that:
  1. Calls `sources.NewFast()` then `srcs.LoadUnicodeSearch()` (search index
     is needed for `unicode_search` and non-exact `unicode_lookup_name`; cost
     is paid once at startup since MCP servers are long-lived).
  2. Calls `mcpserver.NewServer(srcs)` (from `commands/agent/mcpserver/`).
  3. Calls `srv.ServeStdio(context.Background())`.
  4. Exits non-zero on error.
- [ ] Register as a sub-command of `agentCmd` in `init()`.

### B — Tests

**New file: `commands/agent/agenthelp_test.go`**

- [ ] Invoke `agent help` via Cobra's `Execute` with stdout captured to a
  `bytes.Buffer`.
- [ ] Unmarshal JSON output into `AgentHelp`.
- [ ] Assert `Tool == "character"`.
- [ ] Assert `Commands` is non-empty.
- [ ] Find the `named` entry; assert its `Flags` slice contains an entry with
  `Name == "json"` and `Type == "bool"`.
- [ ] Assert no entry with `Name == "agent"` appears at top level (the `agent`
  command is excluded from its own help output).

**New file: `commands/agent/agentexamples_test.go`**

- [ ] Call with no args: assert all six category strings appear at least once
  in the output.
- [ ] Call with arg `"lookup"`: assert all entries have `Category == "lookup"`.
- [ ] Call with arg `"emoji"`: assert all entries have `Category == "emoji"`.
- [ ] Call with arg `"nonexistent"`: assert output is an empty JSON array
  `[]`.

---

## Part C — MCP stdio server

### C.1 — New package `commands/agent/mcpserver`

**New file: `commands/agent/mcpserver/server.go`**

- [ ] Define `type Server struct` wrapping `*mcp.Server` and `*sources.Sources`.
- [ ] Define `func NewServer(srcs *sources.Sources) *Server` that creates the
  `mcp.Server`, registers all eight tools (via helpers in `tools.go`), and
  returns the wrapper.
- [ ] Define `func (s *Server) ServeStdio(ctx context.Context) error` that
  connects via `&mcp.StdioTransport{}`, calls `s.srv.Connect(ctx, transport,
  nil)`, and waits for the session to close.

  ```go
  func (s *Server) ServeStdio(ctx context.Context) error {
      transport := &mcp.StdioTransport{}
      session, err := s.srv.Connect(ctx, transport, nil)
      if err != nil {
          return err
      }
      select {
      case <-ctx.Done():
          session.Close()
          return ctx.Err()
      case <-session.Done():
          return session.Err()
      }
  }
  ```

### C.2 — Character property type

**New file: `commands/agent/mcpserver/charprops.go`**

- [ ] Define output types used by all lookup tools:

  ```go
  type CharProps struct {
      Character       string         `json:"character"`
      Name            string         `json:"name"`
      Hex             string         `json:"hex"`
      Decimal         int            `json:"decimal"`
      UTF8Percent     string         `json:"utf8_percent"`
      UTF8Bytes       string         `json:"utf8_bytes"`
      UTF8Escaped     string         `json:"utf8_escaped"`
      UnicodeEscaped  string         `json:"unicode_escaped"`
      RustEscaped     string         `json:"rust_escaped"`
      JSONEscaped     string         `json:"json_escaped"`
      Block           BlockObj       `json:"block"`
      Category        string         `json:"category"`
      RenderWidth     int            `json:"render_width"`
      HTMLEntities    []string       `json:"html_entities,omitempty"`
      XMLEntities     []string       `json:"xml_entities,omitempty"`
      VimDigraphs     []string       `json:"vim_digraphs,omitempty"`
      X11Digraphs     []string       `json:"x11_digraphs,omitempty"`
      PresentVariants []PresentVar   `json:"presentation_variants,omitempty"`
  }

  type BlockObj struct {
      Name  string `json:"name"`
      Start string `json:"start"`
      End   string `json:"end"`
  }

  type PresentVar struct {
      Selector string `json:"selector"`
      Type     string `json:"type"`
  }
  ```

- [ ] Define `func CharPropsFromRune(r rune, srcs *sources.Sources) CharProps`
  that computes all fields for one codepoint.  This is the single canonical
  place for all field computation, shared by all eight tools.
  - `Character`: `string(r)`.
  - `Name`: `srcs.Unicode.ByRune[r].Name` (empty string if not found).
  - `Hex`: `strings.ToUpper(strconv.FormatUint(uint64(r), 16))`.
  - `Decimal`: `int(r)`.
  - `UTF8Percent`: URL-percent encoding, matching `PRINT_RUNE_UTF8ENC` logic.
  - `UTF8Bytes`: space-separated lowercase two-digit hex bytes.
  - `UTF8Escaped`: `\xNN` per byte, lowercase.
  - `UnicodeEscaped`: `\uXXXX` or `\UXXXXXXXX`.
  - `RustEscaped`: `\u{X…}` no padding.
  - `JSONEscaped`: same logic as `PRINT_RUNE_JSON` in `resultset`.
  - `Block`: from `srcs.UBlocks.LookupInfo(r)`.  If nil, use zero `BlockObj`.
  - `Category`: `unicode.GeneralCategory(r)` (from `unicode/category.go`).
  - `RenderWidth`: `runemanip.DisplayCellWidth(string(r))` first return.
  - `HTMLEntities`, `XMLEntities`: from `entities.HTMLEntitiesReverse`,
    `entities.XMLEntitiesReverse`.
  - `VimDigraphs`, `X11Digraphs`: from `srcs.Vim.DigraphsSliceFor(r)`,
    `srcs.X11.DigraphsSliceFor(r)`.
  - `PresentVariants`: from `unicode.PresentationVariants(r)`, convert selectors
    to `"U+XXXX"` strings.

- [ ] Note: `CharPropsFromRune` must NOT use `resultset.ResultSet` or read
  `resultset.ResultCmdFlags` global state.  It is a pure computation over a
  rune and the sources.

### C.3 — Tool implementations

**New file: `commands/agent/mcpserver/tools.go`**

One exported `RegisterTools(srv *mcp.Server, srcs *sources.Sources)` function
called by `NewServer`.  Each tool uses `mcp.AddTool` with typed
input/output structs.

#### Tool: `unicode_lookup_char`

- [ ] Input: `struct { Char string \`json:"char"\` }`.
- [ ] Validate: `Char` must be exactly one Unicode codepoint; return error if
  empty or more than one rune.
- [ ] Output: `CharProps`.

#### Tool: `unicode_lookup_name`

- [ ] Input: `struct { Name string \`json:"name"\`; Exact bool \`json:"exact"\` }`.
- [ ] If `Exact == true`: look up `strings.ToUpper(Name)` in
  `srcs.Unicode.ByName`.  Return a single-element `[]CharProps` or an error if
  not found.
- [ ] If `Exact == false`: use `srcs.Unicode.Search.Query(Name, -1)` (requires
  search index, loaded at startup in `agentmcp.go`).  Return `[]CharProps`.
- [ ] Output: `[]CharProps`.

#### Tool: `unicode_search`

- [ ] Input: `struct { Query string \`json:"query"\` }`.
- [ ] Uses `srcs.Unicode.Search.Query(Query, -1)`; sort results by codepoint
  (same logic as `named -/`).
- [ ] Output: `[]CharProps`.

#### Tool: `unicode_lookup_codepoint`

- [ ] Input: `struct { Codepoint string \`json:"codepoint"\` }`.
- [ ] Parse `Codepoint`: accept `"U+XXXX"` (hex after `U+`), `"0xXXXX"` (hex
  with `0x` prefix), or bare decimal integer string.
- [ ] Output: `CharProps`.

#### Tool: `unicode_browse_block`

- [ ] Input: `struct { Block string \`json:"block"\` }`.
- [ ] Call `srcs.UBlocks.FindByName(Block)`; return error if not found.
- [ ] Iterate `min..max` runes.  For each `r` in the range, include only
  codepoints that exist in `srcs.Unicode.ByRune` OR (if not present) that fall
  within a known block and can be represented without a name (like CJK).  For
  simplicity in the initial implementation, only emit runes present in
  `srcs.Unicode.ByRune`.
- [ ] Enforce a hard limit of 3000 entries; return an error if the populated
  count would exceed it (consistent with the browse command's `--limit-abort`
  default).
- [ ] Output: `[]CharProps`.

#### Tool: `unicode_list_blocks`

- [ ] No input.
- [ ] Calls `srcs.UBlocks.ListBlocks()`, converts each `BlockInfo` to
  `BlockObj`.
- [ ] Output: `[]BlockObj`.

#### Tool: `unicode_emoji_flag`

- [ ] Input: `struct { CountryCode string \`json:"country_code"\` }`.

- [ ] Define output type:

  ```go
  type FlagResult struct {
      Indicator1 CharProps `json:"indicator_1"`
      Indicator2 CharProps `json:"indicator_2"`
      Combined   string    `json:"combined"`
  }
  ```

- [ ] Validate `CountryCode` is exactly two ASCII letters; convert to uppercase.
- [ ] Map each letter to its Regional Indicator Symbol Letter rune
  (`U+1F1E6 + (letter - 'A')`).
- [ ] Look up both indicators via `CharPropsFromRune`.
- [ ] Set `Combined` to `string(r1) + string(r2)`.
- [ ] Output: `FlagResult`.

#### Tool: `unicode_transform`

**Prerequisite**: The `transform` package functions are currently unexported.
Before implementing this tool, export a minimal API:

- [ ] **Modify `commands/transform/fraktur.go`**: export
  `func TransformFraktur(args []string) (string, error)` wrapping the existing
  anonymous `Transformer`.
- [ ] **Modify `commands/transform/turn.go`**: rename `transformTurn` to
  `TransformTurn` (it is already a named function, just lowercase).
- [ ] **Modify `commands/transform/math.go`**: export
  `func TransformMath(args []string, target string) (string, error)`.  The
  `target` parameter accepts the same values as `--target`; empty string selects
  the default (normal/regular).  Consult the existing `flags.target` logic.
- [ ] **`commands/transform/scream.go`**: `NewEncoder()` and `NewDecoder()` are
  already exported.  No change needed.

MCP tool definition:

- [ ] Input:

  ```go
  type TransformInput struct {
      Type   string `json:"type"   jsonschema:"enum:fraktur,math,scream,scream-decode,turn"`
      Text   string `json:"text"`
      Target string `json:"target,omitempty" jsonschema:"math variant, e.g. bold, italic"`
  }
  ```

- [ ] Output:

  ```go
  type TransformResult struct {
      Input  string `json:"input"`
      Type   string `json:"type"`
      Output string `json:"output"`
  }
  ```

- [ ] Dispatch on `Type`:
  - `"fraktur"` → `transform.TransformFraktur([]string{Text})`
  - `"math"` → `transform.TransformMath([]string{Text}, Target)`
  - `"scream"` → `transform.NewEncoder().Replace(Text)`
  - `"scream-decode"` → `transform.NewDecoder().Replace(Text)`
  - `"turn"` → `transform.TransformTurn([]string{Text})`
  - unknown → return tool error.

### C — Tests

**New file: `commands/agent/mcpserver/charprops_test.go`**

Uses `sources.NewFast()` (and `srcs.LoadUnicodeSearch()` for search tests).

- [ ] `CharPropsFromRune(0x2713, srcs)`:
  - `Name == "CHECK MARK"`.
  - `Hex == "2713"`.
  - `Decimal == 10003`.
  - `UTF8Bytes == "e2 9c 93"`.
  - `UTF8Escaped == "\\xe2\\x9c\\x93"`.
  - `UnicodeEscaped == "\\u2713"`.
  - `RustEscaped == "\\u{2713}"`.
  - `JSONEscaped == "\\u2713"`.
  - `Block.Name == "Dingbats"`, `Block.Start == "U+2700"`.
  - `Category == "So"`.
  - `RenderWidth == 1`.
- [ ] `CharPropsFromRune(0x0041, srcs)`: `Category == "Lu"`, `Name == "LATIN CAPITAL LETTER A"`.
- [ ] `CharPropsFromRune(0x1F600, srcs)`:
  - `UnicodeEscaped == "\\U0001F600"`.
  - `RustEscaped == "\\u{1F600}"`.
  - `UTF8Bytes == "f0 9f 98 80"`.
- [ ] `CharPropsFromRune(0x1F1EB, srcs)`: `Block.Name` contains `"Regional"`.

**New file: `commands/agent/mcpserver/tools_test.go`**

Use `mcp.NewInMemoryTransports()` to test without touching stdio:

```go
func newTestSession(t *testing.T, srcs *sources.Sources) *mcp.ClientSession {
    t.Helper()
    ctx := context.Background()
    srv := mcpserver.NewServer(srcs)
    t1, t2 := mcp.NewInMemoryTransports()
    if _, err := srv.MCP().Connect(ctx, t1, nil); err != nil {
        t.Fatal(err)
    }
    client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
    session, err := client.Connect(ctx, t2, nil)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { session.Close() })
    return session
}
```

(Expose `srv.MCP() *mcp.Server` from the `Server` wrapper for testability.)

- [ ] `unicode_lookup_char` with `char = "✓"`:
  unmarshal result, assert `name == "CHECK MARK"`.
- [ ] `unicode_lookup_char` with `char = ""`: assert `IsError == true` in
  result.
- [ ] `unicode_lookup_char` with `char = "ab"` (two runes): assert `IsError == true`.
- [ ] `unicode_lookup_codepoint` with `codepoint = "U+2713"`:
  assert `name == "CHECK MARK"`.
- [ ] `unicode_lookup_codepoint` with `codepoint = "10003"` (decimal):
  assert `name == "CHECK MARK"`.
- [ ] `unicode_lookup_codepoint` with `codepoint = "0x2713"`:
  assert `name == "CHECK MARK"`.
- [ ] `unicode_lookup_name` with `name = "CHECK MARK"`, `exact = true`:
  assert single result, `name == "CHECK MARK"`.
- [ ] `unicode_lookup_name` with `name = "NONEXISTENT XYZ"`, `exact = true`:
  assert `IsError == true`.
- [ ] `unicode_search` with `query = "snowman"`: assert result array
  non-empty; at least one entry has `name == "SNOWMAN"`.
- [ ] `unicode_list_blocks`: assert result is non-empty; each entry has
  non-empty `name`, `start`, `end`; `start` matches `^U\+[0-9A-F]+$`.
- [ ] `unicode_browse_block` with `block = "Dingbats"`: assert non-empty.
- [ ] `unicode_browse_block` with `block = "Nonexistent Block XYZ"`:
  assert `IsError == true`.
- [ ] `unicode_emoji_flag` with `country_code = "FR"`:
  assert `combined` field non-empty (two regional indicator runes).
- [ ] `unicode_emoji_flag` with `country_code = "fr"` (lowercase):
  same result (case normalisation).
- [ ] `unicode_emoji_flag` with `country_code = "ZZZ"` (three letters):
  assert `IsError == true`.
- [ ] `unicode_transform` with `type = "fraktur"`, `text = "Hello"`:
  assert `output != ""` and `output != input`.
- [ ] `unicode_transform` with `type = "scream-decode"`,
  `text = <result of scream-encoding "Hello">`: assert roundtrip gives `"Hello"`.
- [ ] `unicode_transform` with `type = "invalid"`: assert `IsError == true`.

---

## Implementation order

The items are ordered so each step only depends on work already done.

```
Step 1  unicode/category.go                    (no deps beyond stdlib)
Step 2  util/update_unicode.go  +  regenerate  (no deps beyond existing)
Step 3  unicode/emoji.go: PresentationVariants (needs step 2 generated data)
Step 4  unicode/blocks.go: LookupInfo          (no deps beyond existing)
Step 5  resultset/resultset.go: new JItem      (needs steps 1, 3, 4)

Step 6  commands/agent/agent.go                (needs root package only)
Step 7  commands/agent/agenthelp.go            (needs step 6 + version pkg)
Step 8  commands/agent/agentexamples.go        (needs step 6 only)

Step 9  go get mcp-go-sdk + go mod tidy        (needed before step 10)
Step 10 commands/agent/mcpserver/charprops.go  (needs steps 1, 3, 4)
Step 11 commands/transform: export API         (needed before step 12)
Step 12 commands/agent/mcpserver/tools.go      (needs steps 10, 11)
Step 13 commands/agent/mcpserver/server.go     (needs steps 10, 12)
Step 14 commands/agent/agentmcp.go             (needs steps 6, 13)

Step 15 main.go: blank import for agent pkg    (needs step 6)

Step 16 tests for steps 1–5
Step 17 tests for steps 6–8
Step 18 tests for steps 10–14
```

Steps 1–4 are independent and can proceed in parallel.
Steps 7 and 8 are independent and can proceed in parallel.
Steps 10 and 11 are independent once step 9 is done.

---

## Architectural notes and risks

### `resultset.ResultCmdFlags` is global state

`ResultSet.New` reads `ResultCmdFlags` global flags set by Cobra at parse time.
The MCP server must not use `ResultSet` or `JSONEntry` — those paths are
entangled with CLI flag state.  `CharPropsFromRune` in the `mcpserver` package
is a self-contained computation that avoids this entirely.

### Backward compatibility of `JItem`

All new fields use `omitempty`.  The new block object uses key `"block_info"`
rather than `"block"` so existing consumers reading the `"block"` string are
unaffected.  If a future version wants to replace the string `"block"` field
with a structured object, that is a separate breaking-change decision.

### `unicode_search` startup cost

Ferret's inverted-suffix index is built lazily.  The MCP server calls
`srcs.LoadUnicodeSearch()` eagerly before accepting connections so the latency
hit is at startup, not mid-session.  On current hardware this takes
approximately 100–300 ms.  Document this in `agent mcp --help`.

### Transform package export

The four transformer functions are currently unexported anonymous closures or
package-private named functions.  They must be extracted to exported functions
before they can be called from the `mcpserver` package.  This is a pure
refactor with no change in behaviour; the existing Cobra wiring continues to
call the same underlying logic.  The `scream` package already exports
`NewEncoder`/`NewDecoder` and needs no change.

### MCP content serialisation

The typed `mcp.AddTool` API (`AddTool[In, Out]`) serialises the `Out` value
as JSON text content in `CallToolResult.Content[0]`.  The calling client
receives a `TextContent` with a JSON string.  When the output is a complex
struct (`CharProps`, `[]CharProps`, etc.), clients must parse the `text` field
as JSON.  This is standard MCP practice; document it in `AGENTS.md` under the
MCP section.

### Block name matching in `unicode_browse_block`

`Blocks.FindByName` already performs case-insensitive prefix matching and
returns candidate names on ambiguous input.  The tool should surface candidates
in its error message: `"block 'dingbat' is ambiguous; candidates: Dingbats"`.

### CJK gap in `unicode_browse_block`

CJK Unified Ideographs (U+4E00–U+9FFF) are not in `srcs.Unicode.ByRune`.
`unicode_browse_block` with `block = "CJK Unified Ideographs"` will return
zero results with the initial implementation that only emits runes present in
the map.  This is a known limitation (same as the CLI).  The tool should
include a `note` field in the error or result when the block is a known CJK
block but returns empty.  This can be deferred to a follow-on iteration.
