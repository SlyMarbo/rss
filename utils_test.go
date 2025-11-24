package rss

import (
	"io"
	"strings"
	"testing"
)

func TestReadStringPrefix(t *testing.T) {
	tests := []struct {
		Name   string
		Input  string
		N      int
		Prefix string
	}{
		{
			Name:   "simple",
			Input:  "abcdefg",
			N:      3,
			Prefix: "abc",
		},
		{
			Name:   "zero",
			Input:  "abcdefg",
			N:      0,
			Prefix: "",
		},
		{
			Name:   "full",
			Input:  "abcdefg",
			N:      7,
			Prefix: "abcdefg",
		},
		{
			Name:   "excessive",
			Input:  "abcdefg",
			N:      10,
			Prefix: "abcdefg",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			prefix, full, err := readStringPrefix(strings.NewReader(test.Input), test.N)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if prefix != test.Prefix {
				t.Errorf("incorrect prefix:\nGot:  %q\nWant: %q", prefix, test.Prefix)
			}

			all, err := io.ReadAll(full)
			if err != nil {
				t.Fatalf("failed to read full io.Reader: %v", err)
			}

			if string(all) != test.Input {
				t.Errorf("incorrect full:\nGot:  %q\nWant: %q", all, test.Input)
			}
		})
	}
}
