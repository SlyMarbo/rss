rss
=====

RSS is a small library for simplifying the parsing of RSS and Atom feeds.
The package could do with more testing, but it conforms to the RSS 1.0, 2.0, and Atom 1.0
specifications, to the best of my ability. I've tested it with about 15 different feeds,
and it seems to work fine with them.

If anyone has any problems with feeds being parsed incorrectly, please let me know so that
I can debug and improve the package.

Example usage:
```go
package main

import "github.com/SlyMarbo/rss"

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
	Nickname    string              // This is not set by the package, but could be helpful.
	Title       string
	Description string
	Link        string              // Link to the creator's website.
	UpdateURL   string              // URL of the feed itself.
	Image       *Image              // Feed icon.
	Items       []*Item
	ItemMap     map[string]struct{} // Used in checking whether an item has been seen before.
	Refresh     time.Time           // Earliest time this feed should next be checked.
	Unread      uint32              // Number of unread items. Used by aggregators.
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
