package rss

import (
	"io/ioutil"
	"testing"
)

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
