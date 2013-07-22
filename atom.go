package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

func parseAtom(data []byte, read *db) (*Feed, error) {
	feed := atomFeed{}
	p := xml.NewDecoder(bytes.NewReader(data))
	p.CharsetReader = charsetReader
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}

	out := new(Feed)
	out.Title = feed.Title
	out.Description = feed.Description
	out.Link = feed.Link.Href
	out.Image = feed.Image.Image()
	out.Refresh = time.Now().Add(10 * time.Minute)

	if feed.Items == nil {
		return nil, fmt.Errorf("Error: no feeds found in %q.", string(data))
	}

	out.Items = make([]*Item, 0, len(feed.Items))
	out.ItemMap = make(map[string]struct{})

	// Process items.
	for _, item := range feed.Items {

		// Skip items already known.
		if read.req <- item.ID; <-read.res {
			continue
		}

		next := new(Item)
		next.Title = item.Title
		next.Content = item.Content
		next.Link = item.Link.Href
		if item.Date != "" {
			next.Date, err = parseTime(item.Date)
			if err != nil {
				return nil, err
			}
		}
		next.ID = item.ID
		next.Read = false

		if next.ID == "" {
			fmt.Printf("Warning: Item %q has no ID and will be ignored.\n", next.Title)
			continue
		}

		if _, ok := out.ItemMap[next.ID]; ok {
			fmt.Printf("Warning: Item %q has duplicate ID.\n", next.Title)
			continue
		}

		out.Items = append(out.Items, next)
		out.ItemMap[next.ID] = struct{}{}
		out.Unread++
	}

	return out, nil
}

type atomFeed struct {
	XMLName     xml.Name   `xml:"feed"`
	Title       string     `xml:"title"`
	Description string     `xml:"subtitle"`
	Link        atomLink   `xml:"link"`
	Image       atomImage  `xml:"image"`
	Items       []atomItem `xml:"entry"`
	Updated     string     `xml:"updated"`
}

type atomItem struct {
	XMLName xml.Name `xml:"entry"`
	Title   string   `xml:"title"`
	Content string   `xml:"summary"`
	Link    atomLink `xml:"link"`
	Date    string   `xml:"updated"`
	ID      string   `xml:"id"`
}

type atomImage struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	Url     string   `xml:"url"`
	Height  int      `xml:"height"`
	Width   int      `xml:"width"`
}

type atomLink struct {
	Href string `xml:"href,attr"`
}

func (a *atomImage) Image() *Image {
	out := new(Image)
	out.Title = a.Title
	out.Url = a.Url
	out.Height = uint32(a.Height)
	out.Width = uint32(a.Width)
	return out
}
