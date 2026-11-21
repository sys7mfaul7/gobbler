// Package format holds playlist file format adapters (m3u/m3u8, aimppl4, ...)
// and a small registry so the CLI can pick one by name or file extension.
package format

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sys7mfaul7/gobbler/pkg/playlist"
)

// Reader decodes a playlist from a byte stream.
type Reader interface {
	Read(r io.Reader) (*playlist.Playlist, error)
}

// Writer encodes a playlist to a byte stream.
type Writer interface {
	Write(w io.Writer, p *playlist.Playlist) error
}

// Format describes one playlist format adapter. Read or Write may be nil when
// the adapter only supports one direction.
type Format struct {
	Name       string
	Extensions []string // lower-case, with leading dot
	Reader     Reader
	Writer     Writer
}

var registry = map[string]*Format{}

// Register adds a format adapter. It panics on duplicate names so mistakes
// show up at init time.
func Register(f *Format) {
	if _, dup := registry[f.Name]; dup {
		panic("format: duplicate registration of " + f.Name)
	}
	registry[f.Name] = f
}

// Lookup returns the format registered under name.
func Lookup(name string) (*Format, error) {
	f, ok := registry[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("unknown format %q (known: %s)", name, strings.Join(Names(), ", "))
	}
	return f, nil
}

// ForPath guesses a format from a file name's extension.
func ForPath(path string) (*Format, error) {
	ext := strings.ToLower(filepath.Ext(path))
	for _, f := range registry {
		for _, e := range f.Extensions {
			if e == ext {
				return f, nil
			}
		}
	}
	return nil, fmt.Errorf("no format registered for extension %q", ext)
}

// Names lists registered format names in sorted order.
func Names() []string {
	n := make([]string, 0, len(registry))
	for k := range registry {
		n = append(n, k)
	}
	sort.Strings(n)
	return n
}
