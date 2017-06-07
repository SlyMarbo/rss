package rss

import (
	"fmt"
	"io/ioutil"
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
		data, err := ioutil.ReadFile("testdata/" + test)
		if err != nil {
			t.Fatalf("Reading %s: %v", test, err)
		}

		feed, err := Parse(data)
		if err != nil {
			t.Fatalf("Parsing %s: %v", test, err)
		}

		if len(feed.Items) != want {
			t.Fatalf("%s: expected %q, got %q", test, want, len(feed.Items))
		}
	}
}
func TestParseContent(t *testing.T) {
	tests := map[string]string{
		"rss_2.0_content_encoded": "<p><a href=\"https://example.com/\">Example.com</a> is an example site.</p>",
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

func TestParseItemDateOK(t *testing.T) {
	tests := map[string]string{
		"rss_2.0":                 "2009-09-06 16:45:00 +0000 +0000",
		"rss_2.0_content_encoded": "2009-09-06 16:45:00 +0000 +0000",
		"rss_2.0_enclosure":       "2009-09-06 16:45:00 +0000 +0000",
		"rss_2.0-1":               "2003-06-03 09:39:21 +0000 GMT",
		"rss_2.0-1_enclosure":     "2016-05-14 18:39:34 +0300 +0300",
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

		if fmt.Sprintf("%s", feed.Items[0].Date) != want {
			t.Fatalf("%s: expected %q, got %q", test, want, feed.Items[0].Date)
		}
	}
}

func TestParseItemDateFailure(t *testing.T) {
	tests := map[string]string{
		"rss_2.0": "0001-01-01 00:00:00 +0000 UTC",
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

		if fmt.Sprintf("%s", feed.Items[1].Date) != want {
			t.Fatalf("%s: expected %q, got %q", test, want, feed.Items[1].Date)
		}
		if feed.Items[1].DateValid {
			t.Fatalf("%s: expected %t, got %t", test, false, feed.Items[1].DateValid)
		}
	}
}
