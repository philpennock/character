// Copyright © 2025 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver

import "encoding/json"

// Hand-written JSON Schema objects for each MCP tool's input.
// Only properties that are actually used by the handler are listed.

var schemaLookupChar = json.RawMessage(`{
  "type": "object",
  "properties": {
    "char": {
      "type": "string",
      "description": "A single Unicode character (exactly one codepoint)"
    }
  },
  "required": ["char"],
  "additionalProperties": false
}`)

var schemaLookupName = json.RawMessage(`{
  "type": "object",
  "properties": {
    "name": {
      "type": "string",
      "description": "The Unicode character name to look up"
    },
    "exact": {
      "type": "boolean",
      "description": "If true, require an exact name match; if false, perform substring search",
      "default": false
    }
  },
  "required": ["name"],
  "additionalProperties": false
}`)

var schemaSearch = json.RawMessage(`{
  "type": "object",
  "properties": {
    "query": {
      "type": "string",
      "description": "Substring to search for in Unicode character names"
    }
  },
  "required": ["query"],
  "additionalProperties": false
}`)

var schemaLookupCodepoint = json.RawMessage(`{
  "type": "object",
  "properties": {
    "codepoint": {
      "type": "string",
      "description": "Codepoint in U+XXXX, 0xXXXX, or decimal form (e.g. \"U+2713\", \"10003\")"
    }
  },
  "required": ["codepoint"],
  "additionalProperties": false
}`)

var schemaBrowseBlock = json.RawMessage(`{
  "type": "object",
  "properties": {
    "block": {
      "type": "string",
      "description": "Unicode block name (case-insensitive, partial match accepted)"
    }
  },
  "required": ["block"],
  "additionalProperties": false
}`)

var schemaListBlocks = json.RawMessage(`{
  "type": "object",
  "properties": {},
  "additionalProperties": false
}`)

var schemaEmojiFlag = json.RawMessage(`{
  "type": "object",
  "properties": {
    "country_code": {
      "type": "string",
      "description": "Two-letter ISO 3166-1 alpha-2 country code (e.g. \"FR\", \"JP\")"
    }
  },
  "required": ["country_code"],
  "additionalProperties": false
}`)

var schemaTransform = json.RawMessage(`{
  "type": "object",
  "properties": {
    "type": {
      "type": "string",
      "description": "Transform type: fraktur, math, scream, scream-decode, turn",
      "enum": ["fraktur", "math", "scream", "scream-decode", "turn"]
    },
    "text": {
      "type": "string",
      "description": "Text to transform"
    },
    "target": {
      "type": "string",
      "description": "For math transforms: variant name (bold, italic, frakturnormal, etc.); defaults to normal"
    }
  },
  "required": ["type", "text"],
  "additionalProperties": false
}`)
