# CODE_GUIDE.md — Developer and Agent Guide to the `character` Codebase

This guide explains how the code fits together, provides a recommended reading
order for newcomers, and documents protocols and architectural invariants.  It
is intended for both human contributors and AI coding agents.

For **using** the tool as an AI agent (tool descriptions, example invocations,
MCP tool schemas), see [`AGENTS.md`](AGENTS.md).

(Side-note: I intend to maintain this file via AI, and an AI wrote it.)

---

## Table of Contents

1. [Project Summary](#project-summary)
2. [Reading Order for Comprehension](#reading-order-for-comprehension)
3. [Package Map](#package-map)
4. [Data Flow](#data-flow)
5. [Command Registration Pattern](#command-registration-pattern)
6. [Key Types and Relationships](#key-types-and-relationships)
7. [Generated Code and `go generate`](#generated-code-and-go-generate)
8. [Architectural Invariants and Pitfalls](#architectural-invariants-and-pitfalls)
9. [Protocols](#protocols)
10. [External Dependencies](#external-dependencies)
11. [Testing](#testing)
12. [Build and Run](#build-and-run)

---

## Project Summary

`character` is a Go CLI tool for Unicode codepoint lookup, transformations,
and encoding information.  It is built on [Cobra][cobra] and has no runtime
server — except when run as `character agent mcp`, which starts an [MCP][mcp]
stdio server exposing eight Unicode lookup tools.

License: MIT.  Copyright Phil Pennock.

[cobra]: https://pkg.go.dev/github.com/spf13/cobra
[mcp]: https://modelcontextprotocol.io/

---

## Reading Order for Comprehension

For understanding the codebase end-to-end, read in this order:

|  # | File(s)                                 | Why                                                                 |
| --:| --------------------------------------- | ------------------------------------------------------------------- |
|  1 | `main.go`                               | Entry point; shows all command imports and `go:generate` directives |
|  2 | `commands/root/root.go`                 | Root Cobra command; `AddCommand`, `Start`, `Cobra` exports          |
|  3 | `sources/sources.go`                    | `Sources` struct — the data aggregator everything depends on        |
|  4 | `unicode/unicode.go`                    | `CharInfo`, `Unicode`, `Load`, `LoadSearch`                         |
|  5 | `unicode/blocks.go`                     | `BlockInfo`, `Blocks`, `LookupInfo`, `FindByName`                   |
|  6 | `unicode/category.go`                   | `GeneralCategory(r rune) string`                                    |
|  7 | `unicode/emoji.go`                      | `PresentationVariants`, `Emojiable`                                 |
|  8 | `internal/uformat/uformat.go`           | Pure byte-formatting helpers shared by CLI and MCP output           |
|  9 | `resultset/resultset.go`                | `ResultSet`, `JItem`, JSON rendering — CLI output backbone          |
| 10 | `resultset/cmdrender.go`                | `ResultCmdFlags`, flag registration, `RenderPerCmdline`             |
| 11 | `commands/name/name.go`                 | A simple command — shows the typical command pattern                |
| 12 | `internal/mcpstdio/mcpstdio.go`         | MCP stdio server (~200 lines, hand-rolled)                          |
| 13 | `commands/agent/mcpserver/charprops.go` | `CharProps`, `CharPropsFromRune`                                    |
| 14 | `commands/agent/mcpserver/tools.go`     | Eight MCP tool handlers                                             |
| 15 | `commands/agent/mcpserver/server.go`    | MCP server wiring                                                   |
| 16 | `commands/agent/agentmcp.go`            | `agent mcp` command — ties it all together                          |

After reading these 16 files you will understand every major subsystem.

---

## Package Map

```
github.com/philpennock/character/
│
├── main.go                            Entry point, go:generate directives
├── repo_version.go                    Build-time version info
│
├── commands/                          Cobra command implementations
│   ├── root/                          Root command, AddCommand(), Start()
│   ├── name/                          `name <char>…`  — info about literal characters
│   ├── named/                         `named <NAME>`  — lookup by Unicode name
│   ├── code/                          `code <U+XXXX>` — lookup by codepoint
│   ├── browse/                        `browse -b <block>` — list block contents
│   ├── known/                         `known -b` — list block/charset names
│   ├── aliases/                       `aliases` — alias characters
│   ├── puny/                          `x-puny` — punycode encode/decode
│   ├── region/                        `region <CC>` — flag emoji
│   ├── transform/                     `transform <type>` — fraktur, math, scream, turn
│   ├── version/                       `version` — version info
│   ├── deprecated/                    Deprecated command stubs
│   └── agent/                         Agent sub-command tree
│       ├── agent.go                     Parent `agent` command
│       ├── agenthelp.go                 `agent help` — JSON schema of all commands
│       ├── agentexamples.go             `agent examples` — example invocations
│       ├── agentmcp.go                  `agent mcp` — start MCP stdio server
│       └── mcpserver/                   MCP tool implementations
│           ├── server.go                  Server wrapper + NewServer
│           ├── charprops.go               CharProps struct, CharPropsFromRune
│           ├── schemas.go                 JSON Schema constants for each tool
│           └── tools.go                   registerTools, 8 handler closures
│
├── unicode/                           Unicode data and lookups
│   ├── unicode.go                       CharInfo, Unicode struct, Load/LoadSearch
│   ├── blocks.go                        BlockInfo, Blocks, LookupInfo, FindByName
│   ├── category.go                      GeneralCategory()
│   ├── emoji.go                         PresentationVariants(), Emojiable()
│   ├── regional.go                      Regional indicator helpers
│   ├── sort.go                          Sort interface for CharInfo slices
│   ├── generated_data.go               ← go generate (character name maps)
│   ├── generated_blocks.go             ← go generate (block ranges)
│   └── generated_emoji.go              ← go generate (emojiable/textable sets)
│
├── sources/                           Data source aggregation
│   ├── sources.go                       Sources struct, NewFast(), NewAll()
│   ├── vim.go                           VimDigraph, VimData, digraph loaders
│   ├── x11.go                           X11Data, compose sequence loader
│   ├── generated_static_vim.go         ← go generate
│   └── generated_x11_compose.go        ← go generate
│
├── entities/                          HTML/XML entity lookup
│   ├── generated_html.go              ← go generate (HTMLEntities, reverse map)
│   └── generated_xml.go               ← go generate (XMLEntities, reverse map)
│
├── resultset/                         CLI result rendering
│   ├── resultset.go                     ResultSet, JItem, Add*, PrintJSON, PrintPlain
│   └── cmdrender.go                     ResultCmdFlags, RegisterCmdFlags, RenderPerCmdline
│
├── internal/
│   ├── mcpstdio/                      Hand-rolled MCP stdio server
│   │   └── mcpstdio.go                  Server, Handler, ToolDef, readFrame, writeFrame
│   ├── uformat/                       Pure rune → string formatting helpers
│   │   └── uformat.go                   UTF8Bytes, UTF8Escaped, UnicodeEscaped, etc.
│   ├── runemanip/                     Rune manipulation utilities
│   │   ├── runes.go                     RuneFromHexField
│   │   ├── hexDecode.go                 HexDecodeArgs
│   │   ├── widths.go                    DisplayCellWidth
│   │   ├── regional.go                  Regional indicator helpers
│   │   └── variations.go               Variation selector helpers
│   ├── table/                         Table rendering abstraction
│   │   └── tabular.go                   NewTable, Supported
│   ├── clipboard/                     Clipboard I/O (conditional build)
│   └── encodings/                     Charset decoders
│
├── extra/                             Extra data files, web assets
│
└── util/                              Build-time code generators + tools
    ├── update_unicode.go                Generate unicode/generated_*.go
    ├── update_entities.go               Generate entities/generated_*.go
    ├── update_x11_compose.go            Generate sources/generated_x11_compose.go
    ├── update_static_vim               Bash: generate sources/generated_static_vim.go
    └── mcp_test_driver                 Python: interactive MCP REPL for testing
```

---

## Data Flow

### CLI lookup (e.g. `character name -J ✓`)

```
main.go → root.Start()
  → Cobra dispatch → commands/name.Run
    → sources.NewFast()           load all static data (~1ms)
    → resultset.NewResultSet(srcs)
    → rs.AddCharacterByRune(r)    populate from Sources
    → rs.RenderPerCmdline()       dispatch on ResultCmdFlags
      → rs.PrintJSON()            marshal JItem structs
```

### MCP server (e.g. `character agent mcp`)

```
main.go → root.Start()
  → Cobra dispatch → commands/agent/agentmcp.Run
    → sources.NewFast()
    → srcs.LoadUnicodeSearch()    build Ferret index (~100-300ms)
    → mcpserver.NewServer(srcs)
      → mcpstdio.NewServer("character", version)
      → registerTools(srv, srcs)  register 8 tool handlers
    → srv.ServeStdio(ctx)         read stdin, dispatch, write stdout
      → readFrame (newline-delimited JSON)
      → dispatch on method: initialize | tools/list | tools/call | …
      → handler(ctx, args) → CharPropsFromRune(r, srcs)
      → writeFrame (JSON + \n)
```

### Sources loading

```
sources.NewFast()
  = NewEmpty()
      .LoadUnicode()          unicode.Load() → ByRune, ByName maps
      .LoadUnicodeBlocks()    unicode.LoadBlocks() → sorted []BlockInfo
      .LoadStaticVim()        compiled-in vim digraphs
      .LoadStaticX11()        compiled-in X11 compose sequences

sources.NewAll() additionally calls:
      .LoadUnicodeSearch()    Ferret inverted-suffix index (~100-300ms)
      .LoadLiveVim()          runs `vim` subprocess for live digraphs
```

---

## Command Registration Pattern

Every command package self-registers in `init()`:

```go
// commands/name/name.go
func init() {
    root.AddCommand(nameCmd)
}
```

`main.go` imports each command package with a blank import:

```go
import (
    _ "github.com/philpennock/character/commands/name"
    _ "github.com/philpennock/character/commands/named"
    // …
)
```

This means `main.go` is the single list of enabled commands.  Sub-commands
(e.g. `agent help`, `agent mcp`) are wired within their parent package's
`init()`.

Most commands register shared output flags via:

```go
resultset.RegisterCmdFlags(cmd, supportsOneline)
```

This adds `-v`, `-N`, `-J`, `-1`, `-c`, emoji/text bias flags, etc., and
makes them mutually exclusive where needed.

---

## Key Types and Relationships

```
Sources                             (sources/sources.go)
  ├─ Unicode  unicode.Unicode       map[rune]CharInfo, map[string]CharInfo, Search
  ├─ UBlocks  unicode.Blocks        sorted []BlockInfo
  ├─ Vim      sources.VimData       map[rune][]VimDigraph
  └─ X11      sources.X11Data       map[rune]string

CharInfo                            (unicode/unicode.go)
  {Number rune, Name string, NameWidth int}

BlockInfo                           (unicode/blocks.go)
  {Min rune, Max rune, ID BlockID, Name string}

ResultSet                           (resultset/resultset.go)
  sources *Sources
  items []charItem → JItem (JSON rendering)

JItem                               (resultset/resultset.go)
  CLI JSON output — display-oriented fields, string decimal,
  "block" as string + "block_info" as object

CharProps                           (commands/agent/mcpserver/charprops.go)
  MCP JSON output — structured fields, int decimal,
  "block" as object, no display-oriented fields
  Computed by CharPropsFromRune(r, srcs)
```

`JItem` and `CharProps` are deliberately separate types with different JSON
contracts.  Both use `internal/uformat` for shared byte-formatting.

---

## Generated Code and `go generate`

Run `go generate ./...` from the repo root to regenerate all static data.

| Generator | Input | Output |
|-----------|-------|--------|
| `util/update_unicode.go`        | `unicode/UnicodeData.txt`, `unicode/Blocks.txt`, `unicode/emoji-variation-sequences.txt` | `unicode/generated_data.go`, `unicode/generated_blocks.go`, `unicode/generated_emoji.go` |
| `util/update_entities.go`       | HTML/XML entity specs             | `entities/generated_html.go`, `entities/generated_xml.go` |
| `util/update_x11_compose.go`    | `sources/Compose.en_US.UTF-8.txt` | `sources/generated_x11_compose.go` |
| `util/update_static_vim` (bash) | Runs `vim`                        | `sources/generated_static_vim.go`  |

Generated files are committed to the repository and should not be
hand-edited.  After modifying a generator, re-run `go generate` and commit
the regenerated output alongside the generator change.

---

## Architectural Invariants and Pitfalls

### `ResultCmdFlags` is global mutable state

`resultset.ResultCmdFlags` is a package-level struct populated by Cobra flag
parsing.  It drives CLI rendering decisions.

**Invariant:** `CharPropsFromRune` (MCP path) must never read
`ResultCmdFlags`.  The MCP server does not import `resultset` at all; this is
enforced by Go's package dependency graph.

### `JItem` backward compatibility

New fields in `JItem` use `omitempty`.  The block object uses JSON key
`"block_info"` to coexist with the legacy `"block"` string field.  In
`CharProps` (MCP), there is no legacy, so `"block"` is the structured object.

### `uformat` is the single source of truth for byte-formatting

Both `resultset.JSONEntry` and `mcpserver.CharPropsFromRune` delegate
byte-formatting to `internal/uformat`.  Do not duplicate these computations.

### MCP search index startup cost

`srcs.LoadUnicodeSearch()` builds a [Ferret][ferret] inverted-suffix index,
taking ~100–300 ms.  `agent mcp` calls it eagerly at startup because MCP
servers are long-lived and first-request latency matters more than startup
latency.

[ferret]: https://github.com/argusdusty/Ferret

### CJK Unified Ideographs gap

CJK Unified Ideographs (U+4E00–U+9FFF) are absent from `srcs.Unicode.ByRune`
because the Unicode standard does not assign individual names to that range.
`unicode_browse_block("CJK Unified Ideographs")` returns zero results.  This
matches the CLI behaviour and is not a bug.

### JSON-RPC `id` field type

JSON-RPC 2.0 allows `id` to be a string, number, or null.  The MCP server
preserves the raw `id` as `json.RawMessage` and echoes it unchanged.  Do not
unmarshal it into `interface{}` and re-marshal — that risks converting bare
integers to floats via Go's default JSON decoder.

---

## Protocols

This section documents wire protocols used or referenced by the codebase,
with summaries and pointers to authoritative specifications.

### MCP — Model Context Protocol (stdio transport)

**Used by:** `character agent mcp` → `internal/mcpstdio`

MCP is a protocol for exposing tools to AI agents.  The `character` tool
implements a **tool-only MCP server** over the **stdio transport**.

#### Wire format

MCP stdio uses **newline-delimited JSON** (NDJSON).  Each message is a single
JSON-RPC 2.0 object on one line terminated by `\n`.  Messages MUST NOT
contain embedded newlines.

```
→ {"jsonrpc":"2.0","id":1,"method":"initialize","params":{...}}\n
← {"jsonrpc":"2.0","id":1,"result":{...}}\n
```

**This is NOT the same as LSP framing** (see below).

#### Required methods (tool-only server)

| Method                      | Direction       | Response                                                       |
| --------------------------- | --------------- | -------------------------------------------------------------- |
| `initialize`                | client → server | `InitializeResult` (capabilities, serverInfo, protocolVersion) |
| `notifications/initialized` | client → server | None (notification, no `id`)                                   |
| `tools/list`                | client → server | `{"tools": [...]}`                                             |
| `tools/call`                | client → server | `{"content": [...], "isError": bool}`                          |

#### References

- **Specification:** <https://spec.modelcontextprotocol.io/>
  - Transports: <https://spec.modelcontextprotocol.io/specification/basic/transports/>
  - Stdio transport: newline-delimited, UTF-8, no embedded newlines
- **Protocol version used:** `"2024-11-05"`
- **Website:** <https://modelcontextprotocol.io/>
- **TypeScript SDK (reference impl):** <https://github.com/modelcontextprotocol/typescript-sdk>
- **Go SDK (not used here):** <https://github.com/modelcontextprotocol/go-sdk>

### LSP — Language Server Protocol

**Not used** by this project, but referenced here because the MCP stdio
transport is frequently confused with LSP's wire format.

LSP uses **Content-Length framing** over stdio:

```
Content-Length: 52\r\n
\r\n
{"jsonrpc":"2.0","id":1,"method":"initialize",...}
```

Each message is preceded by HTTP-style headers (`Content-Length: N\r\n\r\n`),
then exactly N bytes of body.  The body may contain newlines.

**Key difference from MCP:** LSP uses `Content-Length` + `\r\n\r\n`; MCP uses
bare `\n`-delimited lines.  Do not mix them up.

#### References

- **Specification:** <https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/>
- **Base protocol (framing):** <https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#baseProtocol>

### JSON-RPC 2.0

Both MCP and LSP build on JSON-RPC 2.0.

- Requests have `jsonrpc`, `method`, `id`, and optional `params`.
- Notifications have `jsonrpc` and `method` but **no `id`** — no response is
  sent.
- Responses have `jsonrpc`, `id`, and either `result` or `error`.
- Error objects have `code` (integer) and `message` (string).
- Standard error codes: `-32700` (parse error), `-32600` (invalid request),
  `-32601` (method not found), `-32602` (invalid params), `-32603` (internal
  error).

#### References

- **Specification:** <https://www.jsonrpc.org/specification>

---

## External Dependencies

| Module                           | Purpose                | Why                                                         |
| -------------------------------- | ---------------------- | ----------------------------------------------------------- |
| `github.com/spf13/cobra`         | CLI framework          | Command tree, flag parsing, help generation                 |
| `github.com/spf13/pflag`         | POSIX flag parsing     | Cobra dependency; also used directly for flag introspection |
| `github.com/argusdusty/Ferret`   | Inverted-suffix index  | Substring search over ~35k Unicode character names          |
| `github.com/atotto/clipboard`    | Clipboard I/O          | `-c` flag: copy result to clipboard                         |
| `github.com/mattn/go-runewidth`  | Terminal cell width    | Correct column alignment in table output                    |
| `github.com/mattn/go-shellwords` | Shell word splitting   | `--argv` flag: re-parse arguments                           |
| `go.pennock.tech/tabular`        | Table rendering        | `-v` verbose table output                                   |
| `golang.org/x/net`               | IDN / punycode         | `x-puny` command                                            |
| `golang.org/x/text`              | Unicode normalisation  | NFC/NFD handling                                            |
| `github.com/liquidgecka/testlib` | Test assertion helpers | Test-only                                                   |

No MCP SDK is used; `internal/mcpstdio` is ~200 lines of hand-rolled code.

---

## Testing

```sh
go test ./...                           # all tests
go test ./internal/mcpstdio/            # MCP protocol tests
go test ./commands/agent/mcpserver/     # MCP tool handler tests
go test ./internal/uformat/            # formatting helper tests
go test ./unicode/                     # unicode data tests
go test ./resultset/                   # CLI rendering tests
```

The `util/mcp_test_driver` script is a Python REPL for interactively testing
the MCP server end-to-end.  It builds the binary, performs the MCP handshake,
and lets you call tools from a prompt:

```sh
./util/mcp_test_driver
mcp> list
mcp> unicode_lookup_char char=✓
mcp> unicode_search query=snowman
```

---

## Build and Run

```sh
# Build
go build -o character .

# Run
./character name ✓
./character named -Jj CHECK MARK
./character agent mcp                 # start MCP server on stdio

# Regenerate data (after updating Unicode source files or generators)
go generate ./...

# Format
gofmt -w .
```
