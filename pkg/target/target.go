// Package target abstracts the media server that receives playlists
// (Jellyfin first, Subsonic/OpenSubsonic later).
package target

import (
	"context"
	"time"
)

// Item is a library item on the media server that a playlist track can be
// matched against.
type Item struct {
	ID       string
	Path     string // server-side file path, if the server exposes it
	Title    string
	Artist   string
	Album    string
	Duration time.Duration
}

// RemotePlaylist is a playlist as it exists on the server.
type RemotePlaylist struct {
	ID      string
	Name    string
	ItemIDs []string
}

// Target is a media server that owns playlists for a user.
type Target interface {
	// Library returns the audio items visible to the user, for matching.
	Library(ctx context.Context, user string) ([]Item, error)
	// Playlists lists the user's playlists.
	Playlists(ctx context.Context, user string) ([]RemotePlaylist, error)
	// Upsert creates or replaces the named playlist so that it contains
	// exactly itemIDs, in order.
	Upsert(ctx context.Context, user, name string, itemIDs []string) (RemotePlaylist, error)
}
