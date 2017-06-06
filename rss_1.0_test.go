package rss

import (
	"io/ioutil"
	"testing"
)

func TestParseRSS(t *testing.T) {
	tests := map[string]string{
		"rss_1.0": "Golem.de",
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
