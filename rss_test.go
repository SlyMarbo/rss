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
		"rss_1.0":   []*Enclosure{{Url: "http://foo.bar/baz.mp3", Type: "audio/mpeg", Length: 65535}},
		"rss_2.0":   []*Enclosure{{Url: "http://example.com/file.mp3", Type: "audio/mpeg", Length: 65535}},
		"rss_2.0-1": []*Enclosure{{Url: "http://gdb.voanews.com/6C49CA6D-C18D-414D-8A51-2B7042A81010_cx0_cy29_cw0_w800_h450.jpg", Type: "image/jpeg", Length: 3123}},
		"atom_1.0":  []*Enclosure{{Url: "http://example.org/audio.mp3", Type: "audio/mpeg", Length: 1234}},
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
