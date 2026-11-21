package redact

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	in := "stream http://host:8096/rest/stream.view?u=shawn&t=abc123&s=salty&id=42 failed"
	out := String(in)
	for _, leak := range []string{"abc123", "salty"} {
		if strings.Contains(out, leak) {
			t.Fatalf("leaked %q in %q", leak, out)
		}
	}
	if !strings.Contains(out, "id=42") || !strings.HasSuffix(out, " failed") {
		t.Fatalf("over-redacted: %q", out)
	}
	if got := String("https://user:pw@example.com/x?api_key=K"); strings.Contains(got, "pw") || strings.Contains(got, "=K") {
		t.Fatalf("leaked: %q", got)
	}
	if got := String(`C:\Music\a.flac`); got != `C:\Music\a.flac` {
		t.Fatalf("changed plain path: %q", got)
	}
}
