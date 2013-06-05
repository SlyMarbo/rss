package rss

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Parse RSS or Atom data.
func Parse(data []byte, seen Seen) (*Feed, error) {

	if seen == nil {
		seen = NewSeen()
	}
	var f *Feed
	var e error
	if strings.Contains(string(data), "<rss") {
		f, e = parseRSS2(data, seen)
	} else if strings.Contains(string(data), "xmlns=\"http://purl.org/rss/1.0/\"") {
		f, e = parseRSS1(data, seen)
	} else {
		f, e = parseAtom(data, seen)
	}
	if f != nil {
		f.Seen = seen
	}
	return f, e
}

// Fetch downloads and parses the RSS feed at the given URL
func Fetch(url string, seen Seen) (*Feed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	out, err := Parse(body, seen)
	if err != nil {
		return nil, err
	}

	if out.Link == "" {
		out.Link = url
	}

	out.UpdateURL = url

	return out, nil
}

// Feed is the top-level structure.
type Feed struct {
	Nickname    string
	Title       string
	Description string
	Link        string
	UpdateURL   string
	Image       *Image
	Items       []*Item
	Seen        Seen
	Refresh     time.Time
	Unread      uint32
}

// Update fetches any new items and updates f.
func (f *Feed) Update() error {

	// Check that we don't update too often.
	if f.Refresh.After(time.Now()) {
		return nil
	}

	if f.UpdateURL == "" {
		return errors.New("Error: feed has no URL.")
	}

	update, err := Fetch(f.UpdateURL, f.Seen)
	if err != nil {
		return err
	}

	f.Refresh = update.Refresh
	f.Title = update.Title
	f.Description = update.Description

	return nil
}

// Item represents a single story.
type Item struct {
	Title      string
	Content    string
	Link       string
	Date       time.Time
	ID         string
	Read       bool
	Authors    []Author
	Categories []string
}

type Author struct {
	Name, Uri, Email string
}

type Image struct {
	Title  string
	Url    string
	Height uint32
	Width  uint32
}
