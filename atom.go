package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

func parseAtom(data []byte) (*Feed, error) {
	warnings := false
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
	for _, link := range feed.Link {
		if link.Rel == "alternate" || link.Rel == "" {
			out.Link = link.Href
			break
		}
	}
	out.Image = feed.Image.Image()
	out.Refresh = time.Now().Add(10 * time.Minute)

	out.Items = make([]*Item, 0, len(feed.Items))
	out.ItemMap = make(map[string]struct{})

	// Process items.
	for _, item := range feed.Items {

		// Skip items already known.
		if _, ok := out.ItemMap[item.ID]; ok {
			continue
		}

		next := new(Item)
		next.Title = item.Title
		next.Summary = item.Summary
		next.Content = item.Content.RAWContent
		if item.Date != "" {
			next.Date, err = parseTime(item.Date)
			if err == nil {
				item.DateValid = true
			}
		}
		next.ID = item.ID
		for _, link := range item.Links {
			if link.Rel == "alternate" || link.Rel == "" {
				next.Link = link.Href
			} else {
				next.Enclosures = append(next.Enclosures, &Enclosure{
					URL:    link.Href,
					Type:   link.Type,
					Length: link.Length,
				})
			}
		}
		next.Read = false

		if next.ID == "" {
			if debug {
				fmt.Printf("[w] Item %q has no ID and will be ignored.\n", next.Title)
				fmt.Printf("[w] %#v\n", item)
			}
			warnings = true
			continue
		}

		if _, ok := out.ItemMap[next.ID]; ok {
			if debug {
				fmt.Printf("[w] Item %q has duplicate ID.\n", next.Title)
				fmt.Printf("[w] %#v\n", next)
			}
			warnings = true
			continue
		}

		out.Items = append(out.Items, next)
		out.ItemMap[next.ID] = struct{}{}
		out.Unread++
	}

	if warnings && debug {
		fmt.Printf("[i] Encountered warnings:\n%s\n", data)
	}

	return out, nil
}

type RAWContent struct {
	RAWContent string `xml:",innerxml"`
}

type atomFeed struct {
	XMLName     xml.Name   `xml:"feed"`
	Title       string     `xml:"title"`
	Description string     `xml:"subtitle"`
	Link        []atomLink `xml:"link"`
	Image       atomImage  `xml:"image"`
	Items       []atomItem `xml:"entry"`
	Updated     string     `xml:"updated"`
}

type atomItem struct {
	XMLName   xml.Name   `xml:"entry"`
	Title     string     `xml:"title"`
	Summary   string     `xml:"summary"`
	Content   RAWContent `xml:"content"`
	Links     []atomLink `xml:"link"`
	Date      string     `xml:"updated"`
	DateValid bool
	ID        string `xml:"id"`
}

type atomImage struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	URL     string   `xml:"url"`
	Height  int      `xml:"height"`
	Width   int      `xml:"width"`
}

type atomLink struct {
	Href   string `xml:"href,attr"`
	Rel    string `xml:"rel,attr"`
	Type   string `xml:"type,attr"`
	Length uint   `xml:"length,attr"`
}

func (a *atomImage) Image() *Image {
	out := new(Image)
	out.Title = a.Title
	out.URL = a.URL
	out.Height = uint32(a.Height)
	out.Width = uint32(a.Width)
	return out
}
