package rss

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestParseRSS(t *testing.T) {
	tests := map[string]string{
		"rss_1.0": "Golem.de",
	}

	for test, want := range tests {
		name := filepath.Join("testdata", test)
		data, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("Reading %s: %v", name, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", name, err)
		}

		if feed.Title != want {
			t.Errorf("%s: got %q, want %q", name, feed.Title, want)
		}

		if len(feed.Items) != 40 {
			t.Errorf("%v: expected 40 items, got: %v", name, len(feed.Items))
		} else {
			for i, item := range feed.Items {
				if !item.DateValid {
					t.Errorf("%v Invalid date for item (#%v): %v", name, i, item.Title)
				}
			}
		}
	}
}
