package rss

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
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
	tests := map[string]Enclosure{
		"rss_1.0":   Enclosure{Url: "http://foo.bar/baz.mp3", Type: "audio/mpeg", Length: 65535},
		"rss_2.0":   Enclosure{Url: "http://example.com/file.mp3", Type: "audio/mpeg", Length: 65535},
		"rss_2.0-1": Enclosure{Url: "http://gdb.voanews.com/6C49CA6D-C18D-414D-8A51-2B7042A81010_cx0_cy29_cw0_w800_h450.jpg", Type: "image/jpeg", Length: 3123},
		"atom_1.0":  Enclosure{Url: "http://example.org/audio.mp3", Type: "audio/mpeg", Length: 1234},
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

		enclosureFound := false
		for _, item := range feed.Items {
			for _, enc := range item.Enclosures {
				enclosureFound = true
				if !reflect.DeepEqual(*enc, want) {
					t.Errorf("%s: expected %#v, got %#v", test, want, *enc)
				}
			}
		}
		if !enclosureFound {
			t.Errorf("No enclosures parsed in test %v", test)
		}
	}
}

func MakeTestdataFetchFunc(file string) FetchFunc {
	return func(url string) (resp *http.Response, err error) {
		// Create mock http.Response
		resp = new(http.Response)
		resp.Body, err = os.Open("testdata/" + file)
		return
	}
}

func TestFeedUnmarshalUpdate(t *testing.T) {
	fetch1 := MakeTestdataFetchFunc("rssupdate-1")
	fetch2 := MakeTestdataFetchFunc("rssupdate-2")
	feed, err := FetchByFunc(fetch1, "http://localhost/dummyrss")
	if err != nil {
		t.Fatalf("Failed fetching testdata 'rssupdate-2': %v", err)
	}

	if 1 != feed.Unread {
		t.Errorf("Expected one unread item initially, got %v", feed.Unread)
	}

	jsonBlob, err := json.Marshal(feed)
	if err != nil {
		t.Fatalf("Failed to marshal Feed %+v\n", feed)
	}

	var unmarshalledFeed Feed
	err = json.Unmarshal(jsonBlob, &unmarshalledFeed)

	var defaultFetchFuncCalled = 0
	DefaultFetchFunc = func(url string) (resp *http.Response, err error) {
		defaultFetchFuncCalled++
		return nil, errors.New("No network in test")
	}

	err = unmarshalledFeed.Update()
	if err != nil {
		t.Logf("Expected failure updating via http in test: %v", err)
	}
	if defaultFetchFuncCalled < 1 {
		t.Error("DefaultFetchFunc was not called during Update()")
	}

	err = unmarshalledFeed.UpdateByFunc(fetch2)
	if err != nil {
		t.Fatalf("Failed updating the feed from testdata 'rssupdate-2': %v", err)
	}

	if 2 != unmarshalledFeed.Unread {
		t.Errorf("Expected two unread items after update, got %v", unmarshalledFeed.Unread)
	}
}

func TestItemGUIDs(t *testing.T) {
	feed1, err := FetchByFunc(MakeTestdataFetchFunc("rss_2.0"), "http://localhost/dummyfeed1")
	if err != nil {
		t.Fatalf("Failed fetching testdata 'rss_2.0': %v", err)
	}

	if len(feed1.Items) != 1 {
		t.Errorf("Expected one item in feed 'rss_2.0', got %v", len(feed1.Items))
	}

	feed2, err := FetchByFunc(MakeTestdataFetchFunc("rssupdate-1"), "http://localhost/dummyfeed2")
	if err != nil {
		t.Fatalf("Failed fetching testdata 'rssupdate-1': %v", err)
	}

	if len(feed2.Items) != 1 {
		t.Errorf("Expected one item in feed 'rssupdate' after step 1, got %v", len(feed2.Items))
	}

	err = feed2.UpdateByFunc(MakeTestdataFetchFunc("rssupdate-2"))
	if err != nil {
		t.Fatalf("Failed fetching testdata 'rssupdate-2': %v", err)
	}

	// rssupdate-2 contains two items, one new item and one old item from rssupdate-1
	if len(feed2.Items) != 2 {
		t.Errorf("Expected two items in feed 'rssupdate' after step 2, got %v", len(feed2.Items))
	}
}
