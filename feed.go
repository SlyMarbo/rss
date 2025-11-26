package rss

import (
	"time"
)

type (
	// Feed is the top-level structure for an
	// RSS/Atom feed.
	Feed struct {
		Title       string
		Language    string
		Author      *Author
		Description string
		Links       []*Link // Link to the creator's website.
		UpdateURL   string  // URL of the feed itself.
		Image       *Image  // Optional feed icon.
		Categories  []string
		Items       []*Item
		NextUpdate  time.Time // Earliest time this feed should next be checked.
		Unread      uint32    // Number of unread items. Used by aggregators.
	}

	// Author records information about
	// a feed's author.
	Author struct {
		Name  string
		URL   string
		Email string
		Links []*Link
	}

	// Link records information about a URL.
	Link struct {
		Href string
		Rel  string
		Type string
	}

	// Image references an image.
	Image struct {
		Title  string
		Href   string
		URL    string
		Height uint32
		Width  uint32
	}

	// Item represents a single story.
	Item struct {
		Title      string
		Summary    string
		Content    string
		Categories []string
		Links      []*Link
		Date       time.Time
		Image      *Image
		ID         string
		Enclosures []*Enclosure

		// When items are stored, they are
		// marked as unread (Read is false).
		// This can be used by aggregators
		// to track when the user has read
		// the item.
		Read bool
	}

	// Enclosure references an external link.
	Enclosure struct {
		URL    string
		Type   string
		Length uint
	}
)
