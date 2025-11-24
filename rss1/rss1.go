// Package rss1 provides parsing functionality for [RSS 1.0] feeds.
//
// [RSS 1.0]: https://web.resource.org/rss/1.0/spec
package rss1

import (
	"encoding/xml"
	"fmt"
	"io"

	"golang.org/x/net/html/charset"
)

// Parse decodes data from r as an RSS 1.0 feed.
func Parse(r io.Reader) (*Feed, error) {
	var feed Feed
	p := xml.NewDecoder(r)
	p.CharsetReader = charset.NewReaderLabel
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}

	if feed.Channel == nil {
		return nil, fmt.Errorf("no channel found in RSS 1.0 feed")
	}

	return &feed, nil
}

type (
	Feed struct {
		XMLName xml.Name `xml:"RDF"`
		Channel *Channel `xml:"channel"`
		Items   []Item   `xml:"item"`
	}

	Channel struct {
		XMLName     xml.Name `xml:"channel"`
		Title       string   `xml:"title"`
		Description string   `xml:"description"`
		Link        string   `xml:"link"`
		Image       *Image   `xml:"image"`
		MinsToLive  int      `xml:"ttl"`
		SkipHours   []int    `xml:"skipHours>hour"`
		SkipDays    []string `xml:"skipDays>day"`
	}

	Image struct {
		XMLName xml.Name `xml:"image"`
		Title   string   `xml:"title"`
		URL     string   `xml:"url"`
		Height  int      `xml:"height"`
		Width   int      `xml:"width"`
	}

	Item struct {
		XMLName     xml.Name    `xml:"item"`
		Title       string      `xml:"title"`
		Description string      `xml:"description"`
		Content     string      `xml:"encoded"`
		Link        string      `xml:"link"`
		PubDate     string      `xml:"pubDate"`
		Date        string      `xml:"date"`
		ID          string      `xml:"guid"`
		Enclosures  []Enclosure `xml:"enclosure"`
	}

	Enclosure struct {
		XMLName xml.Name `xml:"enclosure"`
		URL     string   `xml:"resource,attr"`
		Type    string   `xml:"type,attr"`
		Length  uint     `xml:"length,attr"`
	}
)
