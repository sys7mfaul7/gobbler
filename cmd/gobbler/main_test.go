package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, stdin string, args ...string) (string, string, error) {
	t.Helper()
	var out, errb bytes.Buffer
	err := run(args, strings.NewReader(stdin), &out, &errb)
	return out.String(), errb.String(), err
}

const fixture = "../../pkg/format/testdata/roadtrip.aimppl4"

func TestConvertAIMPToM3U8(t *testing.T) {
	out, _, err := runCLI(t, "", "convert", "--to", "m3u8", fixture)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(out, "#EXTM3U\n#PLAYLIST:Road Trip ☕\n") {
		t.Fatalf("unexpected:\n%s", out)
	}
}

func TestReadWritePipe(t *testing.T) {
	js, _, err := runCLI(t, "", "read", fixture)
	if err != nil {
		t.Fatal(err)
	}
	out, _, err := runCLI(t, js, "write", "--format", "m3u8")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, "#EXTINF:") != 3 {
		t.Fatalf("unexpected:\n%s", out)
	}
}

func TestMatchRedactsAndReports(t *testing.T) {
	dir := t.TempDir()
	lib := filepath.Join(dir, "lib.json")
	os.WriteFile(lib, []byte(`[{"ID":"x1","Path":"/data/music/a.flac","Artist":"A","Title":"One"}]`), 0o644)
	in := `{"name":"P","tracks":[
	  {"location":"http://h/rest/stream.view?u=me&t=TOPSECRET&s=salt&id=nope","artist":"Z","title":"Q"},
	  {"location":"M:\\Music\\a.flac"}]}`
	out, errOut, err := runCLI(t, in, "match", "--library", lib, "--map", `M:\Music=/data/music`)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out+errOut, "TOPSECRET") || strings.Contains(out+errOut, "salt") {
		t.Fatal("token leaked")
	}
	if !strings.Contains(out, `"item_id": "x1"`) || !strings.Contains(errOut, "1/2 tracks matched") {
		t.Fatalf("out=%s err=%s", out, errOut)
	}
}

func TestUsageErrors(t *testing.T) {
	if _, _, err := runCLI(t, "", "read"); err == nil {
		t.Error("read from stdin without --format should fail")
	}
	if _, _, err := runCLI(t, "", "bogus"); err == nil {
		t.Error("unknown subcommand should fail")
	}
	if out, _, err := runCLI(t, "", "formats"); err != nil || !strings.Contains(out, "aimppl4") {
		t.Errorf("formats: %q %v", out, err)
	}
}
