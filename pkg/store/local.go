package store

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	RegisterScheme("file", func(rest string) (Store, error) {
		return &Local{Root: strings.TrimPrefix(rest, "//")}, nil
	})
}

// Local is a Store backed by a directory on the local filesystem.
type Local struct {
	Root string
	// Extensions limits List to these lower-case extensions; empty means all.
	Extensions []string
}

func (l *Local) List(ctx context.Context) ([]Entry, error) {
	var out []Entry
	err := filepath.WalkDir(l.Root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() || !l.wanted(p) {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(l.Root, p)
		out = append(out, Entry{Name: filepath.ToSlash(rel), ModTime: info.ModTime(), Size: info.Size()})
		return nil
	})
	return out, err
}

func (l *Local) wanted(p string) bool {
	if len(l.Extensions) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(p))
	for _, e := range l.Extensions {
		if e == ext {
			return true
		}
	}
	return false
}

func (l *Local) Open(_ context.Context, name string) (io.ReadCloser, error) {
	return os.Open(l.path(name))
}

// Create writes atomically: data goes to a temp file that is renamed into
// place on Close.
func (l *Local) Create(_ context.Context, name string) (io.WriteCloser, error) {
	dst := l.path(name)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return nil, err
	}
	f, err := os.CreateTemp(filepath.Dir(dst), ".gobbler-*")
	if err != nil {
		return nil, err
	}
	return &atomicFile{File: f, dst: dst}, nil
}

func (l *Local) path(name string) string {
	return filepath.Join(l.Root, filepath.FromSlash(filepath.Clean("/"+name)))
}

type atomicFile struct {
	*os.File
	dst string
}

func (a *atomicFile) Close() error {
	if err := a.File.Close(); err != nil {
		os.Remove(a.Name())
		return err
	}
	return os.Rename(a.Name(), a.dst)
}
