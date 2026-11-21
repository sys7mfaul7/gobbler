package format

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/sys7mfaul7/gobbler/pkg/playlist"
)

func init() {
	Register(&Format{Name: "m3u8", Extensions: []string{".m3u8"}, Reader: M3U{}, Writer: M3U{}})
	Register(&Format{Name: "m3u", Extensions: []string{".m3u"}, Reader: M3U{}, Writer: M3U{}})
}

// M3U reads and writes extended M3U / M3U8 playlists. Output is always UTF-8.
type M3U struct{}

// Read parses an M3U/M3U8 playlist. Supported directives: #EXTM3U,
// #PLAYLIST:<name> and #EXTINF:<seconds>,<artist> - <title>. Other comment
// lines are ignored.
func (M3U) Read(r io.Reader) (*playlist.Playlist, error) {
	text, err := decodeText(r)
	if err != nil {
		return nil, err
	}
	p := &playlist.Playlist{}
	var pending *playlist.Track
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(strings.TrimPrefix(sc.Text(), "\uFEFF"))
		switch {
		case line == "":
			continue
		case strings.HasPrefix(line, "#PLAYLIST:"):
			p.Name = strings.TrimSpace(line[len("#PLAYLIST:"):])
		case strings.HasPrefix(line, "#EXTINF:"):
			t := parseExtinf(line[len("#EXTINF:"):])
			pending = &t
		case strings.HasPrefix(line, "#"):
			continue
		default:
			t := playlist.Track{}
			if pending != nil {
				t = *pending
				pending = nil
			}
			t.Location = line
			p.Tracks = append(p.Tracks, t)
		}
	}
	return p, sc.Err()
}

func parseExtinf(s string) playlist.Track {
	var t playlist.Track
	durPart, info, _ := strings.Cut(s, ",")
	// Strip optional attributes: #EXTINF:123 tvg-id="x",Title
	if i := strings.IndexByte(durPart, ' '); i >= 0 {
		durPart = durPart[:i]
	}
	if secs, err := strconv.ParseFloat(strings.TrimSpace(durPart), 64); err == nil && secs > 0 {
		t.Duration = time.Duration(secs * float64(time.Second))
	}
	info = strings.TrimSpace(info)
	if artist, title, ok := strings.Cut(info, " - "); ok {
		t.Artist, t.Title = strings.TrimSpace(artist), strings.TrimSpace(title)
	} else {
		t.Title = info
	}
	return t
}

// Write emits an extended M3U8 playlist.
func (M3U) Write(w io.Writer, p *playlist.Playlist) error {
	bw := bufio.NewWriter(w)
	fmt.Fprintln(bw, "#EXTM3U")
	if p.Name != "" {
		fmt.Fprintf(bw, "#PLAYLIST:%s\n", oneLine(p.Name))
	}
	for _, t := range p.Tracks {
		if t.Title != "" || t.Artist != "" || t.Duration > 0 {
			secs := -1
			if t.Duration > 0 {
				secs = int((t.Duration + time.Second/2) / time.Second)
			}
			info := oneLine(t.Title)
			if t.Artist != "" {
				info = oneLine(t.Artist) + " - " + info
			}
			fmt.Fprintf(bw, "#EXTINF:%d,%s\n", secs, info)
		}
		fmt.Fprintln(bw, oneLine(t.Location))
	}
	return bw.Flush()
}

func oneLine(s string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(s)
}
