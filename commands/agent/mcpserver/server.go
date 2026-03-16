// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver

import (
	"context"

	"github.com/philpennock/character/commands/version"
	"github.com/philpennock/character/internal/mcpstdio"
	"github.com/philpennock/character/sources"
)

// Server wraps the mcpstdio.Server and holds the data sources.
type Server struct {
	inner *mcpstdio.Server
}

// NewServer creates a Server, registers all tools, and returns it.
func NewServer(srcs *sources.Sources) *Server {
	inner := mcpstdio.NewServer("character", version.VersionString)
	registerTools(inner, srcs)
	return &Server{inner: inner}
}

// ServeStdio starts the MCP server on os.Stdin / os.Stdout.
func (s *Server) ServeStdio(ctx context.Context) error {
	return s.inner.ServeStdio(ctx)
}

// Inner returns the underlying mcpstdio.Server, for use in tests.
func (s *Server) Inner() *mcpstdio.Server {
	return s.inner
}
