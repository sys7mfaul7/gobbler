// Package agent implements the playlist-gobbler-aimp watcher: it notices
// changed AIMP playlists in a folder and mirrors them, converted, into a
// store (a local or Nextcloud-synced folder today, WebDAV later).
//
// It polls instead of using OS file notifications so the same code runs on
// Windows, Linux and Android without cgo.
package agent

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/sys7mfaul7/gobbler/pkg/format"
	"github.com/sys7mfaul7/gobbler/pkg/store"
)

// Agent mirrors playlists from Src to Dst in OutFormat.
type Agent struct {
	Src, Dst  store.Store
	OutFormat string
	Log       func(format string, args ...any)

	seen map[string]time.Time
}

// SyncOnce converts every playlist in Src whose modification time changed
// since the last call. It returns the names it wrote.
func (a *Agent) SyncOnce(ctx context.Context) ([]string, error) {
	if a.seen == nil {
		a.seen = map[string]time.Time{}
	}
	out, err := format.Lookup(a.OutFormat)
	if err != nil {
		return nil, err
	}
	if out.Writer == nil {
		return nil, fmt.Errorf("format %s cannot be written", out.Name)
	}
	entries, err := a.Src.List(ctx)
	if err != nil {
		return nil, err
	}
	var wrote []string
	for _, e := range entries {
		in, err := format.ForPath(e.Name)
		if err != nil || in.Reader == nil {
			continue // not a playlist we understand
		}
		if t, ok := a.seen[e.Name]; ok && t.Equal(e.ModTime) {
			continue
		}
		name := strings.TrimSuffix(e.Name, path.Ext(e.Name)) + out.Extensions[0]
		if err := a.convert(ctx, e.Name, name, in, out); err != nil {
			a.logf("skip %s: %v", e.Name, err)
			continue
		}
		a.seen[e.Name] = e.ModTime
		wrote = append(wrote, name)
	}
	return wrote, nil
}

func (a *Agent) convert(ctx context.Context, src, dst string, in, out *format.Format) error {
	r, err := a.Src.Open(ctx, src)
	if err != nil {
		return err
	}
	p, err := in.Reader.Read(r)
	r.Close()
	if err != nil {
		return err
	}
	if p.Name == "" {
		p.Name = strings.TrimSuffix(path.Base(src), path.Ext(src))
	}
	w, err := a.Dst.Create(ctx, dst)
	if err != nil {
		return err
	}
	if err := out.Writer.Write(w, p); err != nil {
		if c, ok := w.(io.Closer); ok {
			c.Close()
		}
		return err
	}
	return w.Close()
}

// Run calls SyncOnce every interval until ctx is cancelled.
func (a *Agent) Run(ctx context.Context, interval time.Duration) error {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		if names, err := a.SyncOnce(ctx); err != nil {
			a.logf("sync: %v", err)
		} else {
			for _, n := range names {
				a.logf("wrote %s", n)
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
		}
	}
}

func (a *Agent) logf(f string, args ...any) {
	if a.Log != nil {
		a.Log(f, args...)
	}
}
