// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver

import (
	"context"
	"time"

	"github.com/philpennock/character/commands/version"
	"github.com/philpennock/character/internal/mcpstdio"
	"github.com/philpennock/character/sources"
)

// Server wraps the mcpstdio.Server and holds the data sources.
type Server struct {
	inner *mcpstdio.Server
	cache *resultCache
}

// NewServer creates a Server, registers all tools, and returns it.
//
// Its instructions are [Instructions]: true of these tools however they are
// served, so this is the constructor for anything that re-hosts them (an
// HTTP service, another transport, an embedding program).  The character CLI's
// own stdio server uses [NewStdioServer] instead.
//
// searchReady, if non-nil, is a channel that is closed once
// srcs.Unicode.Search has been populated.  Tools that require the search
// index will block on it so that initialize / tools/list respond immediately
// while the index loads in the background.  Pass nil to skip the async path
// (e.g. in tests that pre-load everything synchronously).
func NewServer(srcs *sources.Sources, searchReady <-chan struct{}) *Server {
	inner := mcpstdio.NewServer("character", version.VersionString)
	inner.SetInstructions(Instructions)
	cache := newResultCache(5 * time.Minute)
	registerTools(inner, srcs, searchReady, cache)
	return &Server{inner: inner, cache: cache}
}

// NewStdioServer is [NewServer] for `character agent mcp`: the same tools,
// with instructions that also point at the /character-unicode slash command.
// That command is installed from this source tree into a local Claude Code
// (see README), so the pointer is only meaningful when character itself is
// the locally-launched server.
func NewStdioServer(srcs *sources.Sources, searchReady <-chan struct{}) *Server {
	s := NewServer(srcs, searchReady)
	s.inner.SetInstructions(Instructions + localCommandHint)
	return s
}

// Instructions is the transport-neutral server instructions text, returned
// in the MCP InitializeResult to help clients discover and use the Unicode
// tools.  It must stay true for any client, local or hosted: no references
// to locally-installed commands, files or paths.  Keep it concise: it is
// loaded into every conversation context.
const Instructions = `Unicode character database: lookup, search, browse, and transform.

Use these tools when the user asks about Unicode characters, codepoints, ` +
	`character names or properties, Unicode blocks, text encoding (UTF-8 bytes, ` +
	`escape sequences), country flag emoji, or stylistic text transforms (fraktur, ` +
	`math variants, upside-down text).

Key patterns:
- Broad searches: use detail:"summary" with a limit, then look up individual characters for full properties.
- Pagination: pass the returned cursor to get the next page of results.
- Block names accept partial case-insensitive matches; call unicode_list_blocks first if unsure of exact spelling.
- Property objects include language-specific escape sequences (JSON, Rust, C, URL-encoded) ready for direct insertion into source code.`

// localCommandHint is appended by NewStdioServer only.
const localCommandHint = `

For detailed usage guidance, invoke the /character-unicode slash command if available.`

// ServeStdio starts the MCP server on os.Stdin / os.Stdout.
func (s *Server) ServeStdio(ctx context.Context) error {
	return s.inner.ServeStdio(ctx)
}

// Inner returns the underlying mcpstdio.Server, for use in tests.
func (s *Server) Inner() *mcpstdio.Server {
	return s.inner
}
