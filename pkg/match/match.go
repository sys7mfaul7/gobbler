// Package match resolves playlist tracks to media-server library items.
//
// Strategies run in order and the first hit wins:
//
//  1. Server ID: the track already carries an item id (for example the id=
//     parameter of an AIMP Subsonic stream URL).
//  2. Path map: the client path is rewritten through a prefix table
//     (M:\Music -> /data/music) and compared with server paths.
//  3. Metadata: artist + title (+ album when both sides have it) with the
//     duration within a tolerance (default ±2s).
//
// Tracks that match nothing are reported, never silently dropped.
package match

import (
	"net/url"
	"strings"
	"time"
	"unicode"

	"github.com/sys7mfaul7/gobbler/pkg/playlist"
	"github.com/sys7mfaul7/gobbler/pkg/target"
)

// Method says which strategy produced a match.
type Method string

const (
	ByServerID Method = "server-id"
	ByPath     Method = "path"
	ByMetadata Method = "metadata"
	Unmatched  Method = "unmatched"
)

// Result is the outcome for one track.
type Result struct {
	Index  int            `json:"index"`
	Track  playlist.Track `json:"track"`
	ItemID string         `json:"item_id,omitempty"`
	Method Method         `json:"method"`
}

// PathRule rewrites a client path prefix to a server path prefix.
type PathRule struct{ From, To string }

// PathMap is an ordered list of prefix rewrites; the longest From wins.
type PathMap []PathRule

// Apply rewrites p. Matching is case-insensitive and treats '\' and '/' alike,
// so Windows paths map cleanly onto POSIX server paths.
func (m PathMap) Apply(p string) (string, bool) {
	np := normSlashes(p)
	best := -1
	for i, r := range m {
		from := strings.TrimSuffix(normSlashes(r.From), "/")
		if hasPrefixFold(np, from) && (len(np) == len(from) || np[len(from)] == '/') {
			if best < 0 || len(r.From) > len(m[best].From) {
				best = i
			}
		}
	}
	if best < 0 {
		return np, false
	}
	from := strings.TrimSuffix(normSlashes(m[best].From), "/")
	to := strings.TrimSuffix(normSlashes(m[best].To), "/")
	return to + np[len(from):], true
}

// ParseRule parses "FROM=TO".
func ParseRule(s string) (PathRule, bool) {
	from, to, ok := strings.Cut(s, "=")
	return PathRule{From: from, To: to}, ok && from != ""
}

// Matcher resolves tracks against a library snapshot.
type Matcher struct {
	Paths     PathMap
	Tolerance time.Duration // duration tolerance for metadata matches; 0 means 2s

	byID   map[string]bool
	byPath map[string]string
	byMeta map[string][]target.Item
}

// New indexes items for matching.
func New(items []target.Item, paths PathMap) *Matcher {
	m := &Matcher{Paths: paths, byID: map[string]bool{}, byPath: map[string]string{}, byMeta: map[string][]target.Item{}}
	for _, it := range items {
		m.byID[it.ID] = true
		if it.Path != "" {
			m.byPath[strings.ToLower(normSlashes(it.Path))] = it.ID
		}
		k := metaKey(it.Artist, it.Title)
		m.byMeta[k] = append(m.byMeta[k], it)
	}
	return m
}

// Match resolves every track of p in order.
func (m *Matcher) Match(p *playlist.Playlist) []Result {
	out := make([]Result, len(p.Tracks))
	for i, t := range p.Tracks {
		id, how := m.one(t)
		out[i] = Result{Index: i, Track: t, ItemID: id, Method: how}
	}
	return out
}

func (m *Matcher) one(t playlist.Track) (string, Method) {
	if id := serverID(t); id != "" && m.byID[id] {
		return id, ByServerID
	}
	if t.Location != "" {
		p, _ := m.Paths.Apply(t.Location)
		if id, ok := m.byPath[strings.ToLower(p)]; ok {
			return id, ByPath
		}
	}
	tol := m.Tolerance
	if tol == 0 {
		tol = 2 * time.Second
	}
	var best *target.Item
	for _, it := range m.byMeta[metaKey(t.Artist, t.Title)] {
		it := it
		if t.Title == "" {
			break
		}
		if t.Album != "" && it.Album != "" && fold(t.Album) != fold(it.Album) {
			continue
		}
		if t.Duration > 0 && it.Duration > 0 && absDur(t.Duration-it.Duration) > tol {
			continue
		}
		if best == nil || (t.Duration > 0 && absDur(t.Duration-it.Duration) < absDur(t.Duration-best.Duration)) {
			best = &it
		}
	}
	if best != nil {
		return best.ID, ByMetadata
	}
	return "", Unmatched
}

// serverID returns an explicit item id: Track.ServerID, or the id= query
// parameter of a Subsonic/OpenSubsonic stream URL.
func serverID(t playlist.Track) string {
	if t.ServerID != "" {
		return t.ServerID
	}
	if !strings.Contains(t.Location, "://") {
		return ""
	}
	u, err := url.Parse(t.Location)
	if err != nil || !strings.Contains(u.Path, "/rest/") {
		return ""
	}
	return u.Query().Get("id")
}

func metaKey(artist, title string) string { return fold(artist) + "\x00" + fold(title) }

// fold lower-cases and drops punctuation/space so "AC/DC" == "ac dc".
func fold(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func normSlashes(p string) string { return strings.ReplaceAll(p, `\`, "/") }

func hasPrefixFold(s, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[:len(prefix)], prefix)
}

func absDur(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}
