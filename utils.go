package rss

import (
	"bytes"
	"errors"
	"io"
)

// readStringPrefix reads the given bytes from r, returning the
// prefix (as a string) and a reader containing the full contents
// of r.
//
// readStringPrefix will perform a single read, taking up to n
// bytes.
func readStringPrefix(r io.Reader, n int) (prefix string, full io.Reader, err error) {
	if n == 0 {
		return "", r, nil
	}

	buf := make([]byte, n)
	n, err = r.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		return "", nil, err
	}

	prefix = string(buf[:n])
	front := bytes.NewReader(buf[:n])

	if err != nil {
		return prefix, front, nil
	}

	return prefix, io.MultiReader(front, r), nil
}
