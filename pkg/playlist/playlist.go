// Package playlist defines the format-neutral playlist model that every
// gobbler adapter reads from and writes to.
package playlist

import "time"

// Track is one playlist entry. Only Location is required; everything else is
// optional metadata used for matching.
type Track struct {
	// Location is the path or URL exactly as stored in the source playlist.
	Location    string        `json:"location"`
	Title       string        `json:"title,omitempty"`
	Artist      string        `json:"artist,omitempty"`
	Album       string        `json:"album,omitempty"`
	AlbumArtist string        `json:"album_artist,omitempty"`
	Genre       string        `json:"genre,omitempty"`
	Year        string        `json:"year,omitempty"`
	TrackNo     int           `json:"track_no,omitempty"`
	DiscNo      int           `json:"disc_no,omitempty"`
	Duration    time.Duration `json:"duration,omitempty"`
	Size        int64         `json:"size,omitempty"`
	// ServerID is a media-server item id when one is known (for example a
	// Jellyfin item id embedded in a Subsonic stream URL).
	ServerID string `json:"server_id,omitempty"`
	// Extra holds format-specific fields the core does not interpret.
	Extra map[string]string `json:"extra,omitempty"`
}

// Playlist is an ordered list of tracks with a name.
type Playlist struct {
	Name   string            `json:"name"`
	Tracks []Track           `json:"tracks"`
	Meta   map[string]string `json:"meta,omitempty"`
}

// TotalDuration sums the known durations of all tracks.
func (p *Playlist) TotalDuration() time.Duration {
	var d time.Duration
	for _, t := range p.Tracks {
		d += t.Duration
	}
	return d
}
