package datapoint

import (
	// "bytes"
	"fmt"
	"unicode"
	"unicode/utf8"
)

// Clean is Replace with an empty replacement string.
func Clean(s string) (string, error) {
	return Replace(s, "")
}

// Replace removes characters from s that are invalid for OpenTSDB metric and
// tag values and replaces them.
// See: http://opentsdb.net/docs/build/html/user_guide/writing.html#metrics-and-tags
func Replace(s, replacement string) (string, error) {
	var c string
	replaced := false
	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' || r == '.' || r == '/' {
			c += string(r)
			replaced = false
		} else if !replaced {
			c += replacement
			replaced = true
		}
		s = s[size:]
	}
	if len(c) == 0 {
		return "", fmt.Errorf("clean result is empty")
	}
	return c, nil
}

// MustReplace is like Replace, but returns an empty string on error.
func MustReplace(s, replacement string) string {
	r, err := Replace(s, replacement)
	if err != nil {
		return ""
	}
	return r
}
