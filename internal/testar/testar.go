// Package testar is a small helper for checking test vectors in txtar archives.
package testar

import (
	"bytes"
	"encoding/json"
	"io"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/tools/txtar"
	"rsc.io/diff"
)

// TB is a subset of the `testing.TB` interface.
type TB interface {
	Helper()
	Fatalf(format string, args ...any)
}

// Test tests the given vector.
func Test[Feed any](tb TB, name string, parse func(r io.Reader) (Feed, error), format string) {
	tb.Helper()
	ar, err := txtar.ParseFile(filepath.Join("testdata", name) + ".txtar")
	if err != nil {
		tb.Fatalf("failed to parse test vector: %v", err)
		return
	}

	sectionNames := [...][]string{
		{"URL"},
		{"format"},
		{"XML"},
		{"JSON", "error"},
	}

	if len(ar.Files) != len(sectionNames) {
		tb.Fatalf("malformed test vector: got %d sections, want %d", len(ar.Files), len(sectionNames))
		return
	}

	for i, options := range sectionNames {
		if !slices.Contains(options, ar.Files[i].Name) {
			tb.Fatalf("malformed test vector: got section %d with name %q, want %q", i, ar.Files[i].Name, strings.Join(options, " or "))
			return
		}
	}

	gotFormat := strings.TrimSpace(string(ar.Files[1].Data))
	if gotFormat != format {
		tb.Fatalf("invalid test vector: got format %q, want %q", gotFormat, format)
		return
	}

	r := bytes.NewReader(ar.Files[2].Data)
	feed, err := parse(r)
	Check(tb, ar, feed, err)
}

func Check[Feed any](tb TB, ar *txtar.Archive, feed Feed, err error) {
	if ar.Files[3].Name == "error" {
		want := strings.TrimSpace(string(ar.Files[3].Data))
		if err == nil {
			tb.Fatalf("unexpected success while parsing error test vector, want %s", want)
			return
		}

		e := err.Error()
		if e != want {
			tb.Fatalf("wrong error while parsing error test vector:\ngot:  %s\nwant: %s", e, want)
			return
		}

		// All done.
		return
	}

	if err != nil {
		tb.Fatalf("failed to parse test vector: %v", err)
		return
	}

	wantJSON := strings.TrimSpace(string(ar.Files[3].Data)) + "\n"
	enc, err := json.MarshalIndent(feed, "", "\t")
	if err != nil {
		tb.Fatalf("failed to encode test vector: %v", err)
		return
	}

	gotJSON := strings.TrimSpace(string(enc)) + "\n"
	if gotJSON != wantJSON {
		tb.Fatalf("test vector gave different output:\n%s", diff.Format(wantJSON, gotJSON))
		return
	}
}
