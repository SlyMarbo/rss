package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"
)

func parseRSS1(data []byte) (*Feed, error) {
	warnings := false
	feed := rss1_0Feed{}
	p := xml.NewDecoder(bytes.NewReader(data))
	p.CharsetReader = charsetReader
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}
	if feed.Channel == nil {
		return nil, fmt.Errorf("no channel found in %q", string(data))
	}

	channel := feed.Channel

	out := new(Feed)
	out.Title = channel.Title
	out.Description = channel.Description
	out.Link = channel.Link
	out.Image = channel.Image.Image()
	if channel.MinsToLive != 0 {
		sort.Ints(channel.SkipHours)
		next := time.Now().Add(time.Duration(channel.MinsToLive) * time.Minute)
		for _, hour := range channel.SkipHours {
			if hour == next.Hour() {
				next.Add(time.Duration(60-next.Minute()) * time.Minute)
			}
		}
		trying := true
		for trying {
			trying = false
			for _, day := range channel.SkipDays {
				if strings.Title(day) == next.Weekday().String() {
					next.Add(time.Duration(24-next.Hour()) * time.Hour)
					trying = true
					break
				}
			}
		}

		out.Refresh = next
	}

	if out.Refresh.IsZero() {
		out.Refresh = time.Now().Add(DefaultRefreshInterval)
	}

	out.Items = make([]*Item, 0, len(feed.Items))
	out.ItemMap = make(map[string]struct{})

	// Process items.
	for _, item := range feed.Items {

		if item.ID == "" {
			if item.Link == "" {
				if debug {
					fmt.Printf("[w] Item %q has no ID or link and will be ignored.\n", item.Title)
					fmt.Printf("[w] %#v\n", item)
				}
				warnings = true
				continue
			}
			item.ID = item.Link
		}

		// Skip items already known.
		if _, ok := out.ItemMap[item.ID]; ok {
			continue
		}

		next := new(Item)
		next.Title = item.Title
		next.Summary = item.Description
		next.Content = item.Content
		next.Link = item.Link
		if item.Date != "" {
			next.Date, err = parseTime(item.Date)
			if err == nil {
				next.DateValid = true
			}
		} else if item.PubDate != "" {
			next.Date, err = parseTime(item.PubDate)
			if err == nil {
				next.DateValid = true
			}
		}
		next.ID = item.ID
		if len(item.Enclosures) > 0 {
			next.Enclosures = make([]*Enclosure, len(item.Enclosures))
			for i := range item.Enclosures {
				next.Enclosures[i] = item.Enclosures[i].Enclosure()
			}
		}
		next.Read = false

		out.Items = append(out.Items, next)
		out.ItemMap[next.ID] = struct{}{}
		out.Unread++
	}

	if warnings && debug {
		fmt.Printf("[i] Encountered warnings:\n%s\n", data)
	}

	return out, nil
}

type rss1_0Feed struct {
	XMLName xml.Name       `xml:"RDF"`
	Channel *rss1_0Channel `xml:"channel"`
	Items   []rss1_0Item   `xml:"item"`
}

type rss1_0Channel struct {
	XMLName     xml.Name    `xml:"channel"`
	Title       string      `xml:"title"`
	Description string      `xml:"description"`
	Link        string      `xml:"link"`
	Image       rss1_0Image `xml:"image"`
	MinsToLive  int         `xml:"ttl"`
	SkipHours   []int       `xml:"skipHours>hour"`
	SkipDays    []string    `xml:"skipDays>day"`
}

type rss1_0Item struct {
	XMLName     xml.Name `xml:"item"`
	Title       string   `xml:"title"`
	Description string   `xml:"description"`
	Content     string   `xml:"encoded"`
	Link        string   `xml:"link"`
	PubDate     string   `xml:"pubDate"`
	Date        string   `xml:"date"`
	DateValid   bool
	ID          string            `xml:"guid"`
	Enclosures  []rss1_0Enclosure `xml:"enclosure"`
}

type rss1_0Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	URL     string   `xml:"resource,attr"`
	Type    string   `xml:"type,attr"`
	Length  uint     `xml:"length,attr"`
}

func (r *rss1_0Enclosure) Enclosure() *Enclosure {
	out := new(Enclosure)
	out.URL = r.URL
	out.Type = r.Type
	out.Length = r.Length
	return out
}

type rss1_0Image struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	URL     string   `xml:"url"`
	Height  int      `xml:"height"`
	Width   int      `xml:"width"`
}

func (i *rss1_0Image) Image() *Image {
	out := new(Image)
	out.Title = i.Title
	out.URL = i.URL
	out.Height = uint32(i.Height)
	out.Width = uint32(i.Width)
	return out
}
