package rss

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestParseTitle(t *testing.T) {
	tests := map[string]string{
		"rss_0.92":   "Dave Winer: Grateful Dead",
		"rss_1.0":    "Golem.de",
		"rss_2.0":    "RSS Title",
		"rss_2.0-1":  "Liftoff News",
		"atom_1.0":   "Titel des Weblogs",
		"atom_1.0-1": "Golem.de",
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

func TestEnclosure(t *testing.T) {
	tests := map[string][]*Enclosure{
		"rss_1.0":  []*Enclosure{{Url: "http://foo.bar/baz.mp3", Type: "audio/mpeg", Length: 65535}},
		"rss_2.0":  []*Enclosure{{Url: "http://example.com/file.mp3", Type: "audio/mpeg", Length: 65535}},
		"atom_1.0": []*Enclosure{{Url: "http://example.org/audio.mp3", Type: "audio/mpeg", Length: 1234}},
	}

	for test, want := range tests {
		data, err := ioutil.ReadFile("testdata/" + test + "_enclosure")
		if err != nil {
			t.Fatalf("Reading %s: %v", test, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", test, err)
		}

		for _, item := range feed.Items {
			if !reflect.DeepEqual(item.Enclosures, want) {
				t.Errorf("%s: expected %#v, got %#v", test, want, item.Enclosures)
			}
		}
	}
}
