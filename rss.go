package rss

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Parse RSS or Atom data.
func Parse(data []byte) (*Feed, error) {
	if strings.Contains(string(data), "<rss") {
		return parseRSS2(data, database)
	} else if strings.Contains(string(data), "xmlns=\"http://purl.org/rss/1.0/\"") {
		return parseRSS1(data, database)
	} else {
		return parseAtom(data, database)
	}

	panic("Unreachable.")
}

// Fetch downloads and parses the RSS feed at the given URL
func Fetch(url string) (*Feed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return Parse(body)
}

// Feed is the top-level structure.
type Feed struct {
	Title       string
	Description string
	Link        string
	Image       *Image
	Items       []*Item
	Refresh     time.Time
	Unread      uint32
}

func (f *Feed) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("Feed %q\n\t%q\n\t%q\n\t%s\n\tRefresh at %s\n\tUnread: %d\n\tItems:\n",
		f.Title, f.Description, f.Link, f.Image, f.Refresh.Format("Mon 2 Jan 2006 15:04:05 MST"), f.Unread))
	for _, item := range f.Items {
		buf.WriteString(fmt.Sprintf("\t%s\n", item.Format("\t\t")))
	}
	return buf.String()
}

// Item represents a single story.
type Item struct {
	Title   string
	Content string
	Link    string
	Date    time.Time
	ID      string
	Read    bool
}

func (i *Item) String() string {
	return i.Format("")
}

func (i *Item) Format(s string) string {
	return fmt.Sprintf("Item %q\n\t%s%q\n\t%s%s\n\t%s%q\n\t%sRead: %v\n\t%s%q", i.Title, s, i.Link, s,
		i.Date.Format("Mon 2 Jan 2006 15:04:05 MST"), s, i.ID, s, i.Read, s, i.Content)
}

type Image struct {
	Title  string
	Url    string
	Height uint32
	Width  uint32
}

func (i *Image) String() string {
	return fmt.Sprintf("Image %q", i.Title)
}
