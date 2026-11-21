// Package store abstracts where playlist files live: a local folder, a
// WebDAV/Nextcloud folder, SMB (later), and so on.
package store

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// Entry describes one playlist file in a store.
type Entry struct {
	Name    string    // path relative to the store root, forward slashes
	ModTime time.Time // used for change detection
	Size    int64
}

// Store lists, reads and writes playlist files.
type Store interface {
	List(ctx context.Context) ([]Entry, error)
	Open(ctx context.Context, name string) (io.ReadCloser, error)
	Create(ctx context.Context, name string) (io.WriteCloser, error)
}

// Opener builds a Store from the part of a URI after "scheme:".
type Opener func(rest string) (Store, error)

var openers = map[string]Opener{}

// RegisterScheme makes a store available as "<scheme>:<rest>".
func RegisterScheme(scheme string, o Opener) { openers[scheme] = o }

// Open resolves a store URI such as "file:/home/me/Playlists" or
// "nextcloud:/Playlists". A bare path is treated as "file:".
func Open(uri string) (Store, error) {
	scheme, rest, ok := strings.Cut(uri, ":")
	// Treat Windows drive letters ("C:\...") and bare paths as local files.
	if !ok || len(scheme) == 1 {
		scheme, rest = "file", uri
	}
	o, found := openers[scheme]
	if !found {
		return nil, fmt.Errorf("store: unknown scheme %q", scheme)
	}
	return o(rest)
}
