//go:build tools

// Keeps golang.org/x/mobile in go.mod so `gomobile bind ./mobile` works.
package mobile

import _ "golang.org/x/mobile/bind"
