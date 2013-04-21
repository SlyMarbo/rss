rss
=====

RSS is a small library for simplifying the parsing of RSS and Atom feeds.
The package is currently fairly naive, and requires more testing.

Example:
```go
package main

import (
	"github.com/SlyMarbo/rss"
)

func main() {
	feed, err := rss.Fetch("http://example.com/rss")
	if err != nil {
		// handle error.
	}
	
	// ... Some time later ...
	
	err = feed.Update()
	if err != nil {
		// handle error.
	}
}
```

The output structure is pretty much as you'd expect:
```go
type Feed struct {
	Title       string
	Description string
	Link        string
	Image       *Image
	Items       []*Item
	Refresh     time.Time
	Unread      uint32
}

type Item struct {
	Title   string
	Content string
	Link    string
	Date    time.Time
	ID      string
	Read    bool
}

type Image struct {
	Title   string
	Url     string
	Height  uint32
	Width   uint32
}
```
