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
	Register(&Format{Name: "aimppl4", Extensions: []string{".aimppl4"}, Reader: AIMPPL4{}})
}

// AIMPPL4 reads AIMP's native playlist format (.aimppl4), as written by
// AIMP 4.x and 5.x.
//
// The file is UTF-16LE text split into sections:
//
//	#-----SUMMARY-----#    key=value lines (ID, Name, ContentDuration, ...)
//	#-----SETTINGS-----#   key=value lines (Flags, FormatMainLine, ...)
//	#-----CONTENT-----#    group lines starting with '-' (a folder path),
//	                       followed by pipe-delimited track lines
//
// Track field order (0-based):
//
//	0 Path | 1 Title | 2 Artist | 3 Album | 4 AlbumArtist | 5 Genre | 6 Year |
//	7 TrackNo | 8 DiskNo | 9 Composer | 10 Publisher | 11 Bitrate (kbps) |
//	12 Channels | 13 SampleRate (Hz) | 14 Duration (ms) | 15 Size (bytes) |
//	16 BPM | 17 IsActive | 18 Index | 19 PluginReserved | ...
//
// The reader is deliberately lenient: missing trailing fields are treated as
// empty, unknown sections are skipped, and unparsable numbers become zero.
// Writing aimppl4 is not supported yet.
type AIMPPL4 struct{}

// Field indexes in an aimppl4 track line.
const (
	aPath = iota
	aTitle
	aArtist
	aAlbum
	aAlbumArtist
	aGenre
	aYear
	aTrackNo
	aDiskNo
	aComposer
	aPublisher
	aBitrate
	aChannels
	aSampleRate
	aDurationMs
	aSize
	aBPM
	aIsActive
)

// Read parses an aimppl4 playlist.
func (AIMPPL4) Read(r io.Reader) (*playlist.Playlist, error) {
	text, err := decodeText(r)
	if err != nil {
		return nil, err
	}
	p := &playlist.Playlist{Meta: map[string]string{}}
	section := ""
	sawHeader := false
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 64*1024), 4*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimRight(strings.TrimPrefix(sc.Text(), "\uFEFF"), "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.HasPrefix(line, "#-----") && strings.HasSuffix(line, "-----#") {
			section = strings.Trim(line, "#-")
			sawHeader = true
			continue
		}
		switch section {
		case "SUMMARY", "SETTINGS":
			k, v, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			p.Meta[strings.ToLower(section)+"."+k] = v
			if section == "SUMMARY" && k == "Name" {
				p.Name = v
			}
		case "CONTENT":
			if strings.HasPrefix(line, "-") {
				continue // group header: folder path, informational only
			}
			if !strings.Contains(line, "|") {
				return nil, fmt.Errorf("aimppl4: line %d: expected pipe-delimited track", lineNo)
			}
			p.Tracks = append(p.Tracks, aimpTrack(strings.Split(line, "|")))
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if !sawHeader {
		return nil, fmt.Errorf("aimppl4: no #-----SECTION-----# headers found; not an AIMP 4/5 playlist")
	}
	return p, nil
}

func aimpTrack(f []string) playlist.Track {
	get := func(i int) string {
		if i < len(f) {
			return f[i]
		}
		return ""
	}
	num := func(i int) int64 {
		n, _ := strconv.ParseInt(strings.TrimSpace(get(i)), 10, 64)
		return n
	}
	t := playlist.Track{
		Location:    get(aPath),
		Title:       get(aTitle),
		Artist:      get(aArtist),
		Album:       get(aAlbum),
		AlbumArtist: get(aAlbumArtist),
		Genre:       get(aGenre),
		Year:        get(aYear),
		TrackNo:     int(num(aTrackNo)),
		DiscNo:      int(num(aDiskNo)),
		Duration:    time.Duration(num(aDurationMs)) * time.Millisecond,
		Size:        num(aSize),
	}
	extra := map[string]string{}
	for name, i := range map[string]int{
		"composer": aComposer, "publisher": aPublisher, "bitrate": aBitrate,
		"channels": aChannels, "samplerate": aSampleRate, "bpm": aBPM,
		"isactive": aIsActive,
	} {
		if v := get(i); v != "" {
			extra[name] = v
		}
	}
	if len(extra) > 0 {
		t.Extra = extra
	}
	return t
}
