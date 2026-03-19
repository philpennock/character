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
// searchReady, if non-nil, is a channel that is closed once
// srcs.Unicode.Search has been populated.  Tools that require the search
// index will block on it so that initialize / tools/list respond immediately
// while the index loads in the background.  Pass nil to skip the async path
// (e.g. in tests that pre-load everything synchronously).
func NewServer(srcs *sources.Sources, searchReady <-chan struct{}) *Server {
	inner := mcpstdio.NewServer("character", version.VersionString)
	inner.SetInstructions(serverInstructions)
	cache := newResultCache(5 * time.Minute)
	registerTools(inner, srcs, searchReady, cache)
	return &Server{inner: inner, cache: cache}
}

// serverInstructions is returned in the MCP InitializeResult to help clients
// discover and use the Unicode tools.  Keep this concise: it is loaded into
// every conversation context.
const serverInstructions = `Unicode character database: lookup, search, browse, and transform.

Use these tools when the user asks about Unicode characters, codepoints, ` +
	`character names or properties, Unicode blocks, text encoding (UTF-8 bytes, ` +
	`escape sequences), country flag emoji, or stylistic text transforms (fraktur, ` +
	`math variants, upside-down text).

Key patterns:
- Broad searches: use detail:"summary" with a limit, then look up individual characters for full properties.
- Pagination: pass the returned cursor to get the next page of results.
- Block names accept partial case-insensitive matches; call unicode_list_blocks first if unsure of exact spelling.
- Property objects include language-specific escape sequences (JSON, Rust, C, URL-encoded) ready for direct insertion into source code.

For detailed usage guidance, invoke the /character-unicode skill if available.`

// ServeStdio starts the MCP server on os.Stdin / os.Stdout.
func (s *Server) ServeStdio(ctx context.Context) error {
	return s.inner.ServeStdio(ctx)
}

// Inner returns the underlying mcpstdio.Server, for use in tests.
func (s *Server) Inner() *mcpstdio.Server {
	return s.inner
}
