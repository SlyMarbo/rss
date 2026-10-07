package rss

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestReadRootElement(t *testing.T) {
	tests := []struct {
		name  string
		input string
		root  xml.Name
	}{
		{"RSS", "<rss><channel/></rss>", xml.Name{Local: "rss"}},
		{"Atom", `<feed xmlns="http://www.w3.org/2005/Atom"/>`, xml.Name{Space: "http://www.w3.org/2005/Atom", Local: "feed"}},
		{"long preamble", "<!--" + strings.Repeat("x", 1024) + "--><rss/>", xml.Name{Local: "rss"}},
		{"comments", `<!-- <rss --><?feed type="rss"?><feed/>`, xml.Name{Local: "feed"}},
		{"encoding", `<?xml version="1.0" encoding="ISO-8859-1"?><!--caf` + "\xe9" + "--><rss><title>caf\xe9</title>" + strings.Repeat("x", 8192) + "</rss>", xml.Name{Local: "rss"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, input := range []io.Reader{strings.NewReader(test.input), iotest.OneByteReader(strings.NewReader(test.input))} {
				root, full, err := readRootElement(input)
				if err != nil {
					t.Fatal(err)
				}
				if root != test.root {
					t.Errorf("got root %v, want %v", root, test.root)
				}
				all, err := io.ReadAll(full)
				if err != nil {
					t.Fatal(err)
				}
				if string(all) != test.input {
					t.Error("replayed bytes differ from the original input")
				}
			}
		})
	}
}

func TestReadRootElementErrors(t *testing.T) {
	readError := errors.New("reader failed")
	tests := []struct {
		name  string
		input io.Reader
		want  error
	}{
		{"empty", strings.NewReader(""), io.EOF},
		{"before root", iotest.ErrReader(readError), readError},
		{"with root", &dataErrorReader{data: "<rss/>", err: readError}, readError},
		{"with encoding", &dataErrorReader{data: `<?xml version="1.0" encoding="ISO-8859-1"?><rss/>`, err: readError}, readError},
		{"during root", io.MultiReader(strings.NewReader("<rss"), iotest.ErrReader(readError)), readError},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, _, err := readRootElement(test.input)
			if !errors.Is(err, test.want) {
				t.Errorf("got error %v, want %v", err, test.want)
			}
		})
	}

	t.Run("after root", func(t *testing.T) {
		_, full, err := readRootElement(io.MultiReader(strings.NewReader("<rss/>"), iotest.ErrReader(readError)))
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.ReadAll(full); !errors.Is(err, readError) {
			t.Errorf("got error %v, want %v", err, readError)
		}
	})
}

type dataErrorReader struct {
	data string
	err  error
}

func (r *dataErrorReader) Read(p []byte) (int, error) {
	if r.data == "" {
		return 0, io.EOF
	}
	n := copy(p, r.data)
	r.data = r.data[n:]
	if r.data == "" {
		return n, r.err
	}
	return n, nil
}
