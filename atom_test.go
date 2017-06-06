package rss

import (
	"io/ioutil"
	"testing"
)

func TestParseAtomTitle(t *testing.T) {
	tests := map[string]string{
		"atom_1.0":           "Titel des Weblogs",
		"atom_1.0_enclosure": "Titel des Weblogs",
		"atom_1.0-1":         "Golem.de",
	}

	for test, want := range tests {
		data, err := ioutil.ReadFile("testdata/" + test)
		if err != nil {
			t.Fatalf("Reading %s: %v", test, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", test, err)
		}

		if feed.Title != want {
			t.Fatalf("%s: expected %s, got %s", test, want, feed.Title)
		}
	}
}

func TestParseAtomContent(t *testing.T) {
	tests := map[string]string{
		"atom_1.0":           "Volltext des Weblog-Eintrags",
		"atom_1.0_enclosure": "Volltext des Weblog-Eintrags",
		"atom_1.0-1":         "",
	}

	for test, want := range tests {
		data, err := ioutil.ReadFile("testdata/" + test)
		if err != nil {
			t.Fatalf("Reading %s: %v", test, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", test, err)
		}

		if feed.Items[0].Content != want {
			t.Fatalf("%s: expected %s, got %s", test, want, feed.Items[0].Content)
		}
	}
}
