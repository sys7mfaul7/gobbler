// Package mobile is a small gomobile-friendly facade over gobbler for a
// future Android agent app (gomobile bind -target android ./mobile).
// Only strings, ints and errors cross the boundary.
package mobile

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/sys7mfaul7/gobbler/internal/agent"
	"github.com/sys7mfaul7/gobbler/pkg/format"
	"github.com/sys7mfaul7/gobbler/pkg/store"
)

// Version of the gobbler core.
func Version() string { return "dev" }

// Convert converts playlist text from one format to another, for example
// Convert(data, "aimppl4", "m3u8").
func Convert(data []byte, from, to string) (string, error) {
	in, err := format.Lookup(from)
	if err != nil {
		return "", err
	}
	out, err := format.Lookup(to)
	if err != nil {
		return "", err
	}
	if in.Reader == nil || out.Writer == nil {
		return "", fmt.Errorf("cannot convert %s to %s", from, to)
	}
	p, err := in.Reader.Read(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := out.Writer.Write(&buf, p); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// SyncFolder mirrors playlists from srcDir into dstDir once and returns the
// written names, newline-separated.
func SyncFolder(srcDir, dstDir, outFormat string) (string, error) {
	a := &agent.Agent{Src: &store.Local{Root: srcDir}, Dst: &store.Local{Root: dstDir}, OutFormat: outFormat}
	names, err := a.SyncOnce(context.Background())
	return strings.Join(names, "\n"), err
}
