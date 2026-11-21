package match

import (
	"testing"
	"time"

	"github.com/sys7mfaul7/gobbler/pkg/playlist"
	"github.com/sys7mfaul7/gobbler/pkg/target"
)

var lib = []target.Item{
	{ID: "a1", Path: "/data/music/Artist/Album/01 One.flac", Artist: "Artist", Title: "One", Album: "Album", Duration: 200 * time.Second},
	{ID: "b2", Path: "/data/music/AC-DC/Back/02 Two.flac", Artist: "AC/DC", Title: "Two!", Album: "Back", Duration: 180 * time.Second},
	{ID: "c3", Path: "/data/music/AC-DC/Live/02 Two.flac", Artist: "AC/DC", Title: "Two!", Album: "Live", Duration: 240 * time.Second},
}

func TestPathMapApply(t *testing.T) {
	m := PathMap{{From: `M:\Music`, To: "/data/music"}, {From: `M:\Music\Special`, To: "/special"}}
	cases := map[string]string{
		`M:\Music\x\y.flac`:       "/data/music/x/y.flac",
		`m:\music\x.flac`:         "/data/music/x.flac",
		`M:\Music\Special\z.flac`: "/special/z.flac",
	}
	for in, want := range cases {
		if got, ok := m.Apply(in); !ok || got != want {
			t.Errorf("Apply(%q) = %q, %v; want %q", in, got, ok, want)
		}
	}
	if _, ok := m.Apply(`M:\Musical\x.flac`); ok {
		t.Error("prefix must match on a path boundary")
	}
}

func TestMatchStrategies(t *testing.T) {
	m := New(lib, PathMap{{From: `M:\Music`, To: "/data/music"}})
	p := &playlist.Playlist{Tracks: []playlist.Track{
		{Location: "http://srv:8096/rest/stream.view?u=me&t=SECRET&s=salt&id=a1"},
		{Location: `M:\Music\AC-DC\Back\02 Two.flac`},
		{Location: `Z:\elsewhere\two.mp3`, Artist: "ac dc", Title: "two", Duration: 239 * time.Second},
		{Location: `Z:\elsewhere\two.mp3`, Artist: "AC/DC", Title: "Two!", Duration: 300 * time.Second},
		{Location: `Z:\nothing.mp3`, Artist: "Nobody", Title: "Nothing"},
	}}
	want := []struct {
		id string
		m  Method
	}{{"a1", ByServerID}, {"b2", ByPath}, {"c3", ByMetadata}, {"", Unmatched}, {"", Unmatched}}
	for i, r := range m.Match(p) {
		if r.ItemID != want[i].id || r.Method != want[i].m {
			t.Errorf("track %d: got %q/%s want %q/%s", i, r.ItemID, r.Method, want[i].id, want[i].m)
		}
	}
}
