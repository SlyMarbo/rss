package rss

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestParseItemLen(t *testing.T) {
	tests := map[string]int{
		"rss_2.0":                 2,
		"rss_2.0_content_encoded": 1,
		"rss_2.0_enclosure":       1,
		"rss_2.0-1":               4,
		"rss_2.0-1_enclosure":     1,
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

		if len(feed.Items) != want {
			t.Errorf("%s: got %d, want %d", name, len(feed.Items), want)
		}
	}
}
func TestParseContent(t *testing.T) {
	tests := map[string]string{
		"rss_2.0_content_encoded": "<p><a href=\"https://example.com/\">Example.com</a> is an example site.</p>",
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

		if feed.Items[0].Content != want {
			t.Errorf("%s: got %s, want %s", name, feed.Items[0].Content, want)
		}
	}
}

func TestParseItemDateOK(t *testing.T) {
	tests := map[string]string{
		"rss_2.0":                 "2009-09-06 16:45:00 +0000 +0000",
		"rss_2.0_content_encoded": "2009-09-06 16:45:00 +0000 +0000",
		"rss_2.0_enclosure":       "2009-09-06 16:45:00 +0000 +0000",
		"rss_2.0-1":               "2003-06-03 09:39:21 +0000 GMT",
		"rss_2.0-1_enclosure":     "2016-05-14 18:39:34 +0300 +0300",
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

		if fmt.Sprintf("%s", feed.Items[0].Date) != want {
			t.Errorf("%s: got %q, want %q", name, feed.Items[0].Date, want)
		}
	}
}

func TestParseItemDateFailure(t *testing.T) {
	tests := map[string]string{
		"rss_2.0": "0001-01-01 00:00:00 +0000 UTC",
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

		if fmt.Sprintf("%s", feed.Items[1].Date) != want {
			t.Errorf("%s: got %q, want %q", name, feed.Items[1].Date, want)
		}

		if feed.Items[1].DateValid {
			t.Errorf("%s: got unexpected valid date", name)
		}
	}
}
