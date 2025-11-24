// Package rss2 provides parsing functionality for [RSS 2.0] feeds.
//
// [RSS 2.0]: https://www.rssboard.org/rss-specification
package rss2

import (
	"encoding/xml"
	"fmt"
	"io"

	"golang.org/x/net/html/charset"
)

// Parse decodes data from r as an RSS 2.0 feed.
func Parse(r io.Reader) (*Feed, error) {
	var feed Feed
	p := xml.NewDecoder(r)
	p.CharsetReader = charset.NewReaderLabel
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}

	if feed.Channel == nil {
		return nil, fmt.Errorf("no channel found in RSS 2.0 feed")
	}

	return &feed, nil
}

type (
	Feed struct {
		XMLName xml.Name `xml:"rss"`
		Channel *Channel `xml:"channel"`
	}

	Channel struct {
		XMLName     xml.Name   `xml:"channel"`
		Title       string     `xml:"title"`
		Language    string     `xml:"language"`
		Author      string     `xml:"author"`
		Description string     `xml:"description"`
		Link        []Link     `xml:"link"`
		Image       *Image     `xml:"image"`
		Categories  []Category `xml:"category"`
		Items       []Item     `xml:"item"`
		MinsToLive  int        `xml:"ttl"`
		SkipHours   []int      `xml:"skipHours>hour"`
		SkipDays    []string   `xml:"skipDays>day"`
	}

	Link struct {
		Rel      string `xml:"rel,attr"`
		Href     string `xml:"href,attr"`
		Type     string `xml:"type,attr"`
		Chardata string `xml:",chardata"`
	}

	Image struct {
		XMLName xml.Name `xml:"image"`
		Href    string   `xml:"href,attr"`
		Title   string   `xml:"title"`
		URL     string   `xml:"url"`
		Height  int      `xml:"height"`
		Width   int      `xml:"width"`
	}

	Category struct {
		XMLName xml.Name `xml:"category"`
		Name    string   `xml:"text,attr"`
	}

	Item struct {
		XMLName     xml.Name    `xml:"item"`
		Title       string      `xml:"title"`
		Description string      `xml:"description"`
		Content     string      `xml:"encoded"`
		Categories  []string    `xml:"category"`
		Link        string      `xml:"link"`
		PubDate     string      `xml:"pubDate"`
		Date        string      `xml:"date"`
		Image       *Image      `xml:"image"`
		ID          string      `xml:"guid"`
		Enclosures  []Enclosure `xml:"enclosure"`
	}

	Enclosure struct {
		XMLName xml.Name `xml:"enclosure"`
		URL     string   `xml:"url,attr"`
		Type    string   `xml:"type,attr"`
		Length  uint     `xml:"length,attr"`
	}
)
