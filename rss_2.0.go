package rss

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
	"time"
)

func parseRSS2(data []byte) (*Feed, error) {
	warnings := false
	feed := rss2_0Feed{}
	p := xml.NewDecoder(bytes.NewReader(data))
	p.CharsetReader = charsetReader
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}
	if feed.Channel == nil {
		return nil, fmt.Errorf("Error: no channel found in %q.", string(data))
	}

	channel := feed.Channel

	out := new(Feed)
	out.Title = channel.Title
	out.Description = channel.Description
	for _, link := range channel.Link {
		if link.Rel == "" && link.Type == "" && link.Href == "" && link.Chardata != "" {
			out.Link = link.Chardata
			break
		}
	}
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
		out.Refresh = time.Now().Add(10 * time.Minute)
	}

	if channel.Items == nil {
		return nil, fmt.Errorf("Error: no feeds found in %q.", string(data))
	}

	out.Items = make([]*Item, 0, len(channel.Items))
	out.ItemMap = make(map[string]struct{})

	// Process items.
	for _, item := range channel.Items {

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
		next.Category = item.Category
		next.Link = item.Link
		if item.Date != "" {
			next.Date, err = parseTime(item.Date)
			if err != nil {
				return nil, err
			}
		} else if item.PubDate != "" {
			next.Date, err = parseTime(item.PubDate)
			if err != nil {
				return nil, err
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

type rss2_0Feed struct {
	XMLName xml.Name       `xml:"rss"`
	Channel *rss2_0Channel `xml:"channel"`
}

type rss2_0Channel struct {
	XMLName     xml.Name     `xml:"channel"`
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	Link        []rss2_0Link `xml:"link"`
	Image       rss2_0Image  `xml:"image"`
	Items       []rss2_0Item `xml:"item"`
	MinsToLive  int          `xml:"ttl"`
	SkipHours   []int        `xml:"skipHours>hour"`
	SkipDays    []string     `xml:"skipDays>day"`
}

type rss2_0Link struct {
	Rel      string `xml:"rel,attr"`
	Href     string `xml:"href,attr"`
	Type     string `xml:"type,attr"`
	Chardata string `xml:",chardata"`
}

type rss2_0Item struct {
	XMLName     xml.Name          `xml:"item"`
	Title       string            `xml:"title"`
	Description string            `xml:"description"`
	Content     string            `xml:"encoded"`
	Category    string            `xml:"category"`
	Link        string            `xml:"link"`
	PubDate     string            `xml:"pubDate"`
	Date        string            `xml:"date"`
	ID          string            `xml:"guid"`
	Enclosures  []rss2_0Enclosure `xml:"enclosure"`
}

type rss2_0Enclosure struct {
	XMLName xml.Name `xml:"enclosure"`
	Url     string   `xml:"url,attr"`
	Type    string   `xml:"type,attr"`
	Length  int      `xml:"length,attr"`
}

func (r *rss2_0Enclosure) Enclosure() *Enclosure {
	out := new(Enclosure)
	out.Url = r.Url
	out.Type = r.Type
	out.Length = r.Length
	return out
}

type rss2_0Image struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	Url     string   `xml:"url"`
	Height  int      `xml:"height"`
	Width   int      `xml:"width"`
}

func (i *rss2_0Image) Image() *Image {
	out := new(Image)
	out.Title = i.Title
	out.Url = i.Url
	out.Height = uint32(i.Height)
	out.Width = uint32(i.Width)
	return out
}
