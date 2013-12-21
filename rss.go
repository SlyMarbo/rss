package rss

import (
	"bytes"
	"errors"
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

// CacheParsedItemIDs enables or disable Item.ID caching when parsing feeds.
// Returns whether Item.ID were cached prior to function call.
func CacheParsedItemIDs(flag bool) (didCache bool) {
	didCache = !disabled
	disabled = !flag
	return
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

	out, err := Parse(body)
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
	ItemMap     map[string]struct{}
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

	if f.ItemMap == nil {
		f.ItemMap = make(map[string]struct{})
		for _, item := range f.Items {
			if _, ok := f.ItemMap[item.ID]; !ok {
				f.ItemMap[item.ID] = struct{}{}
			}
		}
	}

	update, err := Fetch(f.UpdateURL)
	if err != nil {
		return err
	}

	f.Refresh = update.Refresh
	f.Title = update.Title
	f.Description = update.Description

	for _, item := range update.Items {
		if _, ok := f.ItemMap[item.ID]; !ok {
			f.Items = append(f.Items, item)
			f.ItemMap[item.ID] = struct{}{}
			f.Unread++
		}
	}

	return nil
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
