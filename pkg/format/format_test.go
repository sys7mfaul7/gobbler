package format

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sys7mfaul7/gobbler/pkg/playlist"
)

func readFixture(t *testing.T, name string, r Reader) *playlist.Playlist {
	t.Helper()
	f, err := os.Open("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	p, err := r.Read(f)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return p
}

func TestAIMPPL4Read(t *testing.T) {
	p := readFixture(t, "roadtrip.aimppl4", AIMPPL4{})
	if p.Name != "Road Trip ☕" {
		t.Errorf("name = %q", p.Name)
	}
	if got := p.Meta["settings.Flags"]; got != "2049" {
		t.Errorf("settings.Flags = %q", got)
	}
	if len(p.Tracks) != 3 {
		t.Fatalf("tracks = %d, want 3", len(p.Tracks))
	}
	t0 := p.Tracks[0]
	if t0.Location != `M:\Music\Example Artist\First Album\01 Opening.flac` ||
		t0.Title != "Opening" || t0.Artist != "Example Artist" || t0.Album != "First Album" ||
		t0.Year != "1999" || t0.TrackNo != 1 || t0.DiscNo != 1 ||
		t0.Duration != 245123*time.Millisecond || t0.Size != 31415926 {
		t.Errorf("track 0 = %+v", t0)
	}
	if t0.Extra["composer"] != "Some Composer" || t0.Extra["bitrate"] != "1011" {
		t.Errorf("track 0 extra = %v", t0.Extra)
	}
	if p.Tracks[1].Title != "Middle Ünïcode" {
		t.Errorf("unicode title = %q", p.Tracks[1].Title)
	}
	// Short line: missing trailing fields must not panic.
	t2 := p.Tracks[2]
	if !strings.HasPrefix(t2.Location, `\\fileserver\`) || t2.Duration != 61500*time.Millisecond || t2.Size != 0 {
		t.Errorf("track 2 = %+v", t2)
	}
}

func TestAIMPPL4RejectsNonAIMP(t *testing.T) {
	if _, err := (AIMPPL4{}).Read(strings.NewReader("#EXTM3U\n/a.mp3\n")); err == nil {
		t.Fatal("expected error for non-aimppl4 input")
	}
}

func TestM3URead(t *testing.T) {
	p := readFixture(t, "mix.m3u8", M3U{})
	if p.Name != "Evening Mix" || len(p.Tracks) != 3 {
		t.Fatalf("got name %q, %d tracks", p.Name, len(p.Tracks))
	}
	if tr := p.Tracks[0]; tr.Artist != "Example Artist" || tr.Title != "Opening" || tr.Duration != 245*time.Second {
		t.Errorf("track 0 = %+v", tr)
	}
	if tr := p.Tracks[1]; tr.Duration != 0 || tr.Title != "No Duration" || tr.Location != "relative/track.mp3" {
		t.Errorf("track 1 = %+v", tr)
	}
	if tr := p.Tracks[2]; tr.Location != "/data/music/bare.ogg" || tr.Title != "" {
		t.Errorf("track 2 = %+v", tr)
	}
}

func TestM3URoundTrip(t *testing.T) {
	in := &playlist.Playlist{Name: "RT", Tracks: []playlist.Track{
		{Location: "/x/1.flac", Artist: "A", Title: "One", Duration: 61400 * time.Millisecond},
		{Location: "/x/2.flac"},
	}}
	var buf bytes.Buffer
	if err := (M3U{}).Write(&buf, in); err != nil {
		t.Fatal(err)
	}
	want := "#EXTM3U\n#PLAYLIST:RT\n#EXTINF:61,A - One\n/x/1.flac\n/x/2.flac\n"
	if buf.String() != want {
		t.Fatalf("got:\n%s\nwant:\n%s", buf.String(), want)
	}
	out, err := (M3U{}).Read(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if out.Name != "RT" || len(out.Tracks) != 2 || out.Tracks[0].Title != "One" || out.Tracks[0].Duration != 61*time.Second {
		t.Fatalf("round trip = %+v", out)
	}
}

func TestAIMPToM3U8(t *testing.T) {
	p := readFixture(t, "roadtrip.aimppl4", AIMPPL4{})
	var buf bytes.Buffer
	if err := (M3U{}).Write(&buf, p); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "#EXTINF:245,Example Artist - Opening\n") {
		t.Errorf("unexpected output:\n%s", buf.String())
	}
}

func TestDecodeText(t *testing.T) {
	cases := map[string][]byte{
		"utf8":     []byte("héllo"),
		"utf8bom":  append([]byte{0xEF, 0xBB, 0xBF}, []byte("héllo")...),
		"utf16le":  {0xFF, 0xFE, 'h', 0, 0xE9, 0, 'l', 0, 'l', 0, 'o', 0},
		"utf16be":  {0xFE, 0xFF, 0, 'h', 0, 0xE9, 0, 'l', 0, 'l', 0, 'o'},
		"le-nobom": {'h', 0, 0xE9, 0, 'l', 0, 'l', 0, 'o', 0},
	}
	for name, b := range cases {
		got, err := decodeText(bytes.NewReader(b))
		if err != nil || got != "héllo" {
			t.Errorf("%s: got %q, %v", name, got, err)
		}
	}
}

func TestRegistry(t *testing.T) {
	for path, want := range map[string]string{"a.M3U8": "m3u8", "b.m3u": "m3u", "c.aimppl4": "aimppl4"} {
		f, err := ForPath(path)
		if err != nil || f.Name != want {
			t.Errorf("ForPath(%s) = %v, %v", path, f, err)
		}
	}
	if _, err := Lookup("nope"); err == nil {
		t.Error("expected error for unknown format")
	}
}
