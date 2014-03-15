package rss

import (
	"errors"
	"io"
	"strings"

	"github.com/axgle/mahonia"
)

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch {
	case isCharsetUTF8(charset):
		return input, nil
	case isCharsetISO88591(charset):
		return newCharsetISO88591(input), nil
	default:
		decoder := mahonia.NewDecoder(charset)
		if decoder == nil {
			goto invalidCharset
		}

		return decoder.NewReader(input), nil
	}

invalidCharset:
	return nil, errors.New("CharsetReader: unexpected charset: " + charset)
}

func isCharset(charset string, names []string) bool {
	charset = strings.ToLower(charset)
	for _, n := range names {
		if charset == strings.ToLower(n) {
			return true
		}
	}
	return false
}

func isCharsetUTF8(charset string) bool {
	names := []string{
		"UTF-8",
		// Default
		"",
	}
	return isCharset(charset, names)
}
