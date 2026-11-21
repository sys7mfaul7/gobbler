// Package redact strips credentials from strings before they are logged.
//
// AIMP's Subsonic plugin stores stream URLs such as
// http://host/rest/stream?u=me&t=<token>&s=<salt>&id=... inside playlists,
// and Jellyfin URLs may carry api_key. Anything that might be logged must go
// through String first.
package redact

import (
	"net/url"
	"regexp"
	"strings"
)

var secretParams = map[string]bool{
	"t": true, "s": true, "p": true, "token": true, "api_key": true,
	"apikey": true, "access_token": true, "password": true, "auth": true,
}

var urlRe = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.-]*://[^\s"'<>|]+`)

// String returns s with secrets in any embedded URLs replaced by "REDACTED".
func String(s string) string {
	return urlRe.ReplaceAllStringFunc(s, url1)
}

func url1(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "[unparsable-url]"
	}
	if u.User != nil {
		u.User = url.User("REDACTED")
	}
	q := u.Query()
	changed := false
	for k := range q {
		if secretParams[strings.ToLower(k)] {
			q.Set(k, "REDACTED")
			changed = true
		}
	}
	if changed {
		u.RawQuery = q.Encode()
	}
	return u.String()
}
