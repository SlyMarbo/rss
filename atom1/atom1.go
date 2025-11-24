// Package atom1 provides parsing functionality for [Atom] feeds.
//
// [Atom]: https://datatracker.ietf.org/doc/html/rfc4287
package atom1

import (
	"encoding/xml"
	"io"

	"golang.org/x/net/html/charset"
)

// Parse decodes data from r as an Atom 1.0 feed.
func Parse(r io.Reader) (*Feed, error) {
	var feed Feed
	p := xml.NewDecoder(r)
	p.CharsetReader = charset.NewReaderLabel
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}

	return &feed, nil
}

type (
	Feed struct {
		XMLName     xml.Name `xml:"feed"`
		Title       string   `xml:"title"`
		Description string   `xml:"subtitle"`
		Author      *Author  `xml:"author"`
		Link        []Link   `xml:"link"`
		Image       *Image   `xml:"image"`
		Items       []Item   `xml:"entry"`
		Updated     string   `xml:"updated"`
	}

	Author struct {
		Name  string `xml:"name"`
		URI   string `xml:"uri"`
		Email string `xml:"email"`
		Links []Link `xml:"link"`
	}

	Link struct {
		Href   string `xml:"href,attr"`
		Rel    string `xml:"rel,attr"`
		Type   string `xml:"type,attr"`
		Length uint   `xml:"length,attr"`
	}

	Image struct {
		XMLName xml.Name `xml:"image"`
		Title   string   `xml:"title"`
		URL     string   `xml:"url"`
		Height  int      `xml:"height"`
		Width   int      `xml:"width"`
	}

	Item struct {
		XMLName xml.Name `xml:"entry"`
		Title   string   `xml:"title"`
		Summary string   `xml:"summary"`
		Content *Content `xml:"content"`
		Links   []Link   `xml:"link"`
		Date    string   `xml:"updated"`
		ID      string   `xml:"id"`
	}

	Content struct {
		RAWContent string `xml:",innerxml"`
	}
)
