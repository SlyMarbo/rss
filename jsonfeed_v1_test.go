package rss

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestParseJSONFeedV1(t *testing.T) {
	tests := map[string]string{
		"jsonfeed_v1": "JSON Feed",
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

		if len(feed.Items) != 1 {
			t.Errorf("%v: expected 1 item, got: %v", name, len(feed.Items))
		} else {
			for i, item := range feed.Items {
				if !item.DateValid {
					t.Errorf("%v Invalid date for item (#%v): %v", name, i, item.Title)
				}
			}
		}
	}
}
