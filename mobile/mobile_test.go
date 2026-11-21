package mobile

import (
	"os"
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	b, err := os.ReadFile("../pkg/format/testdata/roadtrip.aimppl4")
	if err != nil {
		t.Fatal(err)
	}
	out, err := Convert(b, "aimppl4", "m3u8")
	if err != nil || !strings.HasPrefix(out, "#EXTM3U") {
		t.Fatalf("%q %v", out, err)
	}
	if _, err := Convert(b, "aimppl4", "nope"); err == nil {
		t.Fatal("expected error")
	}
}
