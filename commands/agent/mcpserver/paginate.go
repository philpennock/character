// Copyright © 2026 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

package mcpserver

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/philpennock/character/internal/uformat"
	"github.com/philpennock/character/sources"
	"github.com/philpennock/character/unicode"
)

// maxQueryLen is the maximum allowed byte length for search/name/block input
// strings.  The longest real Unicode character name is ~83 characters; 200
// bytes is generous while still preventing bloated cursors and pointless
// searches.
const maxQueryLen = 200

// defaultLimit is the default page size for paginated responses.
const defaultLimit = 200

// --- cursor codec ---

type cursorData struct {
	Type   string `json:"t"`           // "search", "name", "block"
	Query  string `json:"q,omitempty"` // search/name query
	Block  string `json:"b,omitempty"` // canonical block name
	Offset int    `json:"o"`           // offset into result set
}

func encodeCursor(c cursorData) string {
	b, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (cursorData, error) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return cursorData{}, fmt.Errorf("invalid cursor: %w", err)
	}
	var c cursorData
	if err := json.Unmarshal(b, &c); err != nil {
		return cursorData{}, fmt.Errorf("invalid cursor payload: %w", err)
	}
	if c.Type == "" {
		return cursorData{}, fmt.Errorf("invalid cursor: missing type")
	}
	if c.Offset < 0 {
		return cursorData{}, fmt.Errorf("invalid cursor: negative offset")
	}
	return c, nil
}

// --- pagination parameters ---

// pageFields are the shared pagination fields embedded in each pageable
// handler's parameter struct.
type pageFields struct {
	Detail string `json:"detail"`
	Limit  int    `json:"limit"`
	Cursor string `json:"cursor"`
}

type pageParams struct {
	Detail string
	Limit  int
	Cursor string
}

func (pf pageFields) normalized() pageParams {
	pp := pageParams{
		Detail: pf.Detail,
		Limit:  pf.Limit,
		Cursor: pf.Cursor,
	}
	if pp.Detail == "" {
		pp.Detail = "full"
	}
	if pp.Detail != "full" && pp.Detail != "summary" {
		pp.Detail = "full"
	}
	if pp.Limit <= 0 {
		pp.Limit = defaultLimit
	}
	return pp
}

// --- result cache ---

type cacheEntry struct {
	runes   []rune
	created time.Time
}

type resultCache struct {
	mu      sync.Mutex
	entries map[string]cacheEntry
	ttl     time.Duration
}

func newResultCache(ttl time.Duration) *resultCache {
	return &resultCache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
	}
}

func (c *resultCache) get(key string) ([]rune, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok {
		return nil, false
	}
	if time.Since(e.created) > c.ttl {
		delete(c.entries, key)
		return nil, false
	}
	return e.runes, true
}

func (c *resultCache) put(key string, runes []rune) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry{runes: runes, created: time.Now()}
}

// --- response envelope ---

// PageResponse is the paginated response envelope returned by search,
// lookup-name (substring), and browse-block tools.
type PageResponse struct {
	// Full mode
	Results []CharProps `json:"results,omitempty"`
	// Summary mode
	Columns []string   `json:"columns,omitempty"`
	Rows    [][]string `json:"rows,omitempty"`
	// Common
	Count  int    `json:"count"`
	Total  int    `json:"total"`
	Cursor string `json:"cursor,omitempty"`
}

// --- summary row builder ---

var summaryColumns = []string{"character", "codepoint", "name", "category"}

func summaryRow(r rune, srcs *sources.Sources) []string {
	name := ""
	if ci, ok := srcs.Unicode.ByRune[r]; ok {
		name = ci.Name
	}
	return []string{string(r), uformat.Codepoint(r), name, unicode.GeneralCategory(r)}
}

// --- shared paginator ---

// paginateRunes slices allRunes according to pp and cursorBase, builds the
// appropriate response (full or summary), and returns it as a JSON string.
func paginateRunes(
	allRunes []rune,
	pp pageParams,
	cursorBase cursorData,
	srcs *sources.Sources,
) (string, error) {
	total := len(allRunes)
	offset := 0

	if pp.Cursor != "" {
		cd, err := decodeCursor(pp.Cursor)
		if err != nil {
			return "", err
		}
		if cd.Type != cursorBase.Type {
			return "", fmt.Errorf("cursor type %q does not match expected %q; start a new search", cd.Type, cursorBase.Type)
		}
		offset = cd.Offset
	}

	if offset > total {
		offset = total
	}

	end := min(offset+pp.Limit, total)

	page := allRunes[offset:end]

	resp := PageResponse{
		Count: len(page),
		Total: total,
	}

	if end < total {
		next := cursorBase
		next.Offset = end
		resp.Cursor = encodeCursor(next)
	}

	if pp.Detail == "summary" {
		resp.Columns = summaryColumns
		resp.Rows = make([][]string, len(page))
		for i, r := range page {
			resp.Rows[i] = summaryRow(r, srcs)
		}
	} else {
		resp.Results = make([]CharProps, len(page))
		for i, r := range page {
			resp.Results[i] = CharPropsFromRune(r, srcs)
		}
	}

	return jsonResult(resp)
}

// runesFromSearch converts search results into a sorted rune slice.
func runesFromSearch(found []any) []rune {
	cis := charInfoSlice(found)
	runes := make([]rune, len(cis))
	for i, ci := range cis {
		runes[i] = ci.Number
	}
	return runes
}
