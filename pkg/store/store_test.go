package store

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalRoundTrip(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	w, err := s.Create(ctx, "sub/a.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(w, "#EXTM3U\n")
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("x"), 0o644)
	s.(*Local).Extensions = []string{".m3u8"}
	list, err := s.List(ctx)
	if err != nil || len(list) != 1 || list[0].Name != "sub/a.m3u8" {
		t.Fatalf("list = %+v, %v", list, err)
	}
	r, err := s.Open(ctx, "sub/a.m3u8")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	b, _ := io.ReadAll(r)
	if string(b) != "#EXTM3U\n" {
		t.Fatalf("content = %q", b)
	}
}

func TestLocalNoEscape(t *testing.T) {
	l := &Local{Root: "/root/dir"}
	if got := l.path("../../etc/passwd"); got != filepath.FromSlash("/root/dir/etc/passwd") {
		t.Fatalf("path escaped root: %s", got)
	}
}

func TestOpenUnknownScheme(t *testing.T) {
	if _, err := Open("gopher:/x"); err == nil {
		t.Fatal("expected error")
	}
}
