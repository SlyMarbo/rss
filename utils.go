package rss

import (
	"bytes"
	"encoding/xml"
	"io"

	"golang.org/x/net/html/charset"
)

// readRootElement finds the XML root and returns a reader of the original bytes.
func readRootElement(r io.Reader) (xml.Name, io.Reader, error) {
	front := &prefixReader{Reader: r}
	decoder := xml.NewDecoder(front)
	decoder.CharsetReader = charset.NewReaderLabel

	for {
		token, err := decoder.Token()
		if front.err != nil {
			return xml.Name{}, nil, front.err
		}
		if err != nil {
			return xml.Name{}, nil, err
		}
		if root, ok := token.(xml.StartElement); ok {
			return root.Name, io.MultiReader(bytes.NewReader(front.buf.Bytes()), r), nil
		}
	}
}

type prefixReader struct {
	io.Reader
	buf bytes.Buffer
	err error
}

func (r *prefixReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	r.buf.Write(p[:n])
	// The decoder can return a root token before reporting a buffered read error.
	if err != nil && err != io.EOF {
		r.err = err
	}
	return n, err
}
