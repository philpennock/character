// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

// Package mcpserver re-exports the MCP server core from internal/mcpstdio so
// that consumers outside this module can inspect registered tools and re-use
// handlers in alternative transports (e.g. HTTP via the official MCP SDK).
//
// The Server type embeds the internal implementation; this package adds only
// the [ToolRegistration] type and the [Server.Tools] accessor.
package mcpserver

import (
	"context"
	"io"

	"github.com/philpennock/character/internal/mcpstdio"
)

// Handler processes a single tool call.  It receives the tool arguments as
// raw JSON and returns a result string, or an error if the call failed.
//
// This is a type alias for the internal handler signature so that callers
// outside this module can reference it without importing internal packages.
type Handler = mcpstdio.Handler

// ToolDef describes a single MCP tool for registration and for tools/list.
//
// This is a type alias for the internal tool definition so that callers
// outside this module can reference it without importing internal packages.
type ToolDef = mcpstdio.ToolDef

// ToolRegistration pairs a [ToolDef] with its [Handler], allowing external
// consumers to iterate over all registered tools and re-register them on a
// different transport.
type ToolRegistration struct {
	ToolDef
	Handler Handler
}

// Server wraps [mcpstdio.Server] and exposes the registered tools for
// external consumption.  All stdio-serving methods delegate directly.
type Server struct {
	inner *mcpstdio.Server
}

// NewServer creates a Server with the given name and version strings (used in
// the InitializeResult serverInfo).
func NewServer(name, version string) *Server {
	return &Server{inner: mcpstdio.NewServer(name, version)}
}

// SetInstructions sets the instructions string returned in the
// InitializeResult.  Per MCP 2025-03-26 §Lifecycle, clients MAY surface this
// to the model to guide tool discovery and usage.
func (s *Server) SetInstructions(text string) {
	s.inner.SetInstructions(text)
}

// Instructions returns the instructions string, or "" if none was set.
func (s *Server) Instructions() string {
	return s.inner.Instructions()
}

// AddTool registers a tool.  Registration order is preserved in tools/list
// and in the slice returned by [Server.Tools].
func (s *Server) AddTool(def ToolDef, h Handler) {
	s.inner.AddTool(def, h)
}

// Tools returns a snapshot of all registered tools in registration order.
// The returned slice is a copy; callers may safely retain it.
func (s *Server) Tools() []ToolRegistration {
	inner := s.inner.Tools()
	out := make([]ToolRegistration, len(inner))
	for i, t := range inner {
		out[i] = ToolRegistration{
			ToolDef: t.Def,
			Handler: t.Handler,
		}
	}
	return out
}

// Name returns the server name (as reported in InitializeResult serverInfo).
func (s *Server) Name() string { return s.inner.Name() }

// Version returns the server version (as reported in InitializeResult
// serverInfo).
func (s *Server) Version() string { return s.inner.Version() }

// ServeStdio runs the MCP server on os.Stdin / os.Stdout.
func (s *Server) ServeStdio(ctx context.Context) error {
	return s.inner.ServeStdio(ctx)
}

// ServeConn runs the MCP server on the given reader/writer.  It is the
// testable entry point; production code typically calls [Server.ServeStdio].
func (s *Server) ServeConn(ctx context.Context, r io.Reader, w io.Writer) error {
	return s.inner.ServeConn(ctx, r, w)
}

// Inner returns the underlying [mcpstdio.Server].  This is intended for use
// by code within this module that needs direct access; external consumers
// should use [Server.Tools] instead.
func (s *Server) Inner() *mcpstdio.Server {
	return s.inner
}
