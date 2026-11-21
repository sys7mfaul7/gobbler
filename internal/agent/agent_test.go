package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sys7mfaul7/gobbler/pkg/store"
)

func TestSyncOnce(t *testing.T) {
	src, dst := t.TempDir(), t.TempDir()
	fixture, err := os.ReadFile("../../pkg/format/testdata/roadtrip.aimppl4")
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(src, "Road.aimppl4"), fixture, 0o644)
	os.WriteFile(filepath.Join(src, "notes.txt"), []byte("ignore me"), 0o644)
	a := &Agent{Src: &store.Local{Root: src}, Dst: &store.Local{Root: dst}, OutFormat: "m3u8"}
	ctx := context.Background()

	got, err := a.SyncOnce(ctx)
	if err != nil || len(got) != 1 || got[0] != "Road.m3u8" {
		t.Fatalf("first sync = %v, %v", got, err)
	}
	b, _ := os.ReadFile(filepath.Join(dst, "Road.m3u8"))
	if !strings.Contains(string(b), "#PLAYLIST:Road Trip") {
		t.Fatalf("output:\n%s", b)
	}
	if got, _ := a.SyncOnce(ctx); len(got) != 0 {
		t.Fatalf("unchanged file re-synced: %v", got)
	}
	future := time.Now().Add(time.Hour)
	os.Chtimes(filepath.Join(src, "Road.aimppl4"), future, future)
	if got, _ := a.SyncOnce(ctx); len(got) != 1 {
		t.Fatalf("changed file not re-synced: %v", got)
	}
}
