package format

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf16"
	"unicode/utf8"
)

// decodeText reads all of r and returns it as UTF-8, honouring UTF-8,
// UTF-16LE and UTF-16BE byte-order marks. Without a BOM it sniffs for
// UTF-16LE (AIMP's default) by looking for NUL high bytes, else assumes UTF-8.
func decodeText(r io.Reader) (string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	switch {
	case bytes.HasPrefix(b, []byte{0xEF, 0xBB, 0xBF}):
		b = b[3:]
	case bytes.HasPrefix(b, []byte{0xFF, 0xFE}):
		return decodeUTF16(b[2:], false)
	case bytes.HasPrefix(b, []byte{0xFE, 0xFF}):
		return decodeUTF16(b[2:], true)
	case looksUTF16LE(b):
		return decodeUTF16(b, false)
	}
	if !utf8.Valid(b) {
		return "", errors.New("input is not valid UTF-8 or UTF-16")
	}
	return string(b), nil
}

func looksUTF16LE(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	n := len(b)
	if n > 256 {
		n = 256
	}
	zeros := 0
	for i := 1; i < n; i += 2 {
		if b[i] == 0 {
			zeros++
		}
	}
	return zeros*2 >= n/2 // at least half of the odd bytes are NUL
}

func decodeUTF16(b []byte, bigEndian bool) (string, error) {
	if len(b)%2 != 0 {
		return "", errors.New("UTF-16 input has an odd number of bytes")
	}
	u := make([]uint16, len(b)/2)
	for i := range u {
		if bigEndian {
			u[i] = uint16(b[2*i])<<8 | uint16(b[2*i+1])
		} else {
			u[i] = uint16(b[2*i+1])<<8 | uint16(b[2*i])
		}
	}
	return string(utf16.Decode(u)), nil
}
