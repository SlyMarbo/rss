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

The library does its best to follow the appropriate specifications and not to set the Refresh time
too soon. It currently follows all update time management methods in the RSS 1.0, 2.0, and Atom 1.0
specifications. If one is not provided, it defaults to 10 minute intervals. If you are having issues
with feed providors dropping connections, please let me know and I can increase this default, or you
can increase the Refresh time manually. The Feed.Update method uses this Refresh time, so if Update
seems to be returning very quickly with no new items, it's likely not making a request due to the
provider's Refresh interval.

This is seeing thorough use in [RS3][1], but development is still active.


[1]: https://github.com/SlyMarbo/rs3        "RS3"
