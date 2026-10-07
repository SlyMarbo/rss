/*
Package rss is a small library for simplifying the parsing of RSS and Atom feeds.

The package conforms to the [RSS 1.0], [RSS 2.0], and [Atom 1.0] specifications.

If you encounter any problems with feeds being parsed incorrectly, please
open an issue on GitHub.

Example usage:

	package main

	import (
		"context"

		"github.com/SlyMarbo/rss/v2"
	)

	func main() {
		ctx := context.Background()
		reader, err := rss.NewReader()
		if err != nil {
			// handle error.
		}

		feed, err := reader.Fetch(ctx, "https://example.com/rss")
		if err != nil {
			// handle error.
		}

		// ... Some time later ...

		err = reader.UpdatePatiently(ctx, feed)
		if err != nil {
			// handle error.
		}
	}

The library does its best to follow the appropriate specifications and not to set the `NextUpdate` time
too soon. It currently follows all update time management methods in the RSS 1.0, 2.0, and Atom 1.0
specifications. If one is not provided, it defaults to 12 hour intervals. If you are having issues
with feed providors dropping connections, use `WithDefaultUpdateInterval` to set the default update
interval.

The project is not proactively maintained, but I'll respond to issues and PRs as soon as I can.

[Atom 1.0]: https://datatracker.ietf.org/doc/html/rfc4287
[RSS 1.0]: https://web.resource.org/rss/1.0/spec
[RSS 2.0]: https://www.rssboard.org/rss-specification
*/
package rss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"
)

// ErrTooSoon indicates that a feed cannot be updated
// yet, as the desired interval between updates has
// not yet elapsed.
var ErrTooSoon = errors.New("rss: too soon to update feed")

// Reader stores the configuration data used to fetch,
// parse, and update RSS feeds. Reader supports RSS 2.0,
// RSS 1.0, and Atom 1.0. A single Reader can process
// many different feeds.
type Reader struct {
	// General config.
	time func() time.Time // Optional override on the current time.
	slog *slog.Logger

	// Networking config.
	http *http.Client

	// Parsing config.
	defaultUpdateInterval        time.Duration // Minimum time before updating a feed that hasn't specified an interval.
	timeLayouts                  []string
	timeLayoutsWithNamedLocation []string
}

// NewReader is used to configure a [Reader]. The
// resulting reader will start with sensible defaults,
// which are then updated by any provided options.
//
// NewReader will only return an error if one of the
// provided options fails.
func NewReader(options ...Option) (*Reader, error) {
	r := &Reader{
		// General config.
		slog: slog.New(slog.DiscardHandler), // Default to no logging.

		// Networking config.
		http: http.DefaultClient,

		// Parsing config.
		defaultUpdateInterval:        12 * time.Hour,
		timeLayouts:                  slices.Clone(defaultTimeLayouts),
		timeLayoutsWithNamedLocation: slices.Clone(defaultTimeLayoutsWithNamedLocation),
	}

	for _, option := range options {
		err := option(r)
		if err != nil {
			return nil, err
		}
	}

	return r, nil
}

// now returns the current time.
func (r *Reader) now() time.Time {
	if r.time == nil {
		return time.Now()
	}

	return r.time()
}

// debug records the given log message at the debug
// level.
func (r *Reader) debug(ctx context.Context, msg string, attrs ...slog.Attr) {
	r.slog.LogAttrs(ctx, slog.LevelDebug, msg, attrs...)
}

// info records the given log message at the info
// level.
func (r *Reader) info(ctx context.Context, msg string, attrs ...slog.Attr) {
	r.slog.LogAttrs(ctx, slog.LevelInfo, msg, attrs...)
}

// warn records the given log message at the warn
// level.
func (r *Reader) warn(ctx context.Context, msg string, attrs ...slog.Attr) {
	r.slog.LogAttrs(ctx, slog.LevelWarn, msg, attrs...)
}

// error records the given log message at the error
// level.
func (r *Reader) error(ctx context.Context, msg string, attrs ...slog.Attr) {
	r.slog.LogAttrs(ctx, slog.LevelError, msg, attrs...)
}

// Fetch downloads and parses the RSS/Atom feed at
// the given URL.
func (r *Reader) Fetch(ctx context.Context, url string) (*Feed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct GET request: %v", err)
	}

	res, err := r.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request feed: %v", err)
	}

	feed, err := r.Parse(url, res.Body)
	if err != nil {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		return nil, fmt.Errorf("failed to parse feed: %v", err)
	}

	err = res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to close response body: %v", err)
	}

	return feed, nil
}

// Parse RSS or Atom data.
func (r *Reader) Parse(url string, data io.Reader) (*Feed, error) {
	root, data, err := readRootElement(data)
	if err != nil {
		return nil, err
	}

	switch root.Local {
	case "rss":
		return r.parseRSS2(url, data)
	case "RDF":
		return r.parseRSS1(url, data)
	default:
		return r.parseAtom1(url, data)
	}
}

// Update attempts to update the feed with new items.
// If it is too soon to update the feed, Update will
// return immediately with [ErrTooSoon].
//
// Callers that would rather wait until an update can
// be performed should instead use [Reader.UpdatePatiently].
func (r *Reader) Update(ctx context.Context, feed *Feed) error {
	if feed.UpdateURL == "" {
		return errors.New("feed has no update URL")
	}

	// Check whether enough time has elapsed since
	// the last update.
	if feed.NextUpdate.After(r.now()) {
		return ErrTooSoon
	}

	return r.update(ctx, feed)
}

// UpdatePatiently attempts to update the feed with
// new items. If it is too soon to update the feed,
// UpdatePatiently will wait until the chosen interval
// has elapsed and then attempt an update. If the
// given context expires while UpdatePatiently is
// waiting, it will return with the context's error.
//
// Callers that would rather return immediately if it
// is too soon to perform an update should instead use
// [Reader.Update].
func (r *Reader) UpdatePatiently(ctx context.Context, feed *Feed) error {
	if feed.UpdateURL == "" {
		return errors.New("feed has no update URL")
	}

	err := r.waitForFeed(ctx, feed)
	if err != nil {
		return err
	}

	return r.update(ctx, feed)
}

func (r *Reader) waitForFeed(ctx context.Context, feed *Feed) error {
	// Check whether enough time has elapsed since
	// the last update.
	if now := r.now(); feed.NextUpdate.After(now) {
		interval := feed.NextUpdate.Sub(now)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			return nil
		}
	}

	return nil
}

// update performs an update to a feed. It assumes
// an appropriate interval has passed since the last
// update.
func (r *Reader) update(ctx context.Context, feed *Feed) error {
	seen := make(map[string]struct{})
	for _, item := range feed.Items {
		seen[item.ID] = struct{}{}
	}

	update, err := r.Fetch(ctx, feed.UpdateURL)
	if err != nil {
		return fmt.Errorf("failed to update feed: %v", err)
	}

	feed.NextUpdate = update.NextUpdate
	feed.Title = update.Title
	feed.Description = update.Description

	for _, item := range update.Items {
		if _, ok := seen[item.ID]; !ok {
			seen[item.ID] = struct{}{}
			feed.Items = append(feed.Items, item)
			feed.Unread++
		}
	}

	return nil
}

// Option represents a function that can be used to
// configure a [Reader]. Options are typically used
// with [NewReader].
type Option func(*Reader) error

// General config.

// WithNow configures the reader to use the given
// callback to determine the current time.
//
// By default, readers use [time.Now].
func WithNow(now func() time.Time) Option {
	return func(r *Reader) error {
		r.time = now
		return nil
	}
}

// WithLogger configures the reader to use the given
// structured logger.
func WithLogger(logger *slog.Logger) Option {
	return func(r *Reader) error {
		r.slog = logger
		return nil
	}
}

// Networking config.

// WithHTTPClient configures the reader to use the given
// HTTP client when fetching or updating feeds.
func WithHTTPClient(client *http.Client) Option {
	return func(r *Reader) error {
		r.http = client
		return nil
	}
}

// Parsing config.

// WithDefaultUpdateInterval configures the reader
// to use the given minimum interval between updates
// if the feed has not specified its interval.
func WithDefaultUpdateInterval(interval time.Duration) Option {
	return func(r *Reader) error {
		r.defaultUpdateInterval = interval
		return nil
	}
}

// WithExtraTimeLayouts configures the reader to
// use the given time layouts in addition to the
// built-in layouts when tyring to parse timestamps.
//
// These layouts must not use names like `"EST"` to
// represent time zones. If the layouts do use named
// locations, use [WithExtraTimeLayoutsWithNamedLocation]
// instead.
func WithExtraTimeLayouts(layouts ...string) Option {
	return func(r *Reader) error {
		// Check these don't use named layouts.
		for _, layout := range layouts {
			if strings.Contains(layout, "MST") {
				return fmt.Errorf("cannot use layout %q with WithExtraTimeLayouts: has named layout", layout)
			}
		}

		r.timeLayouts = append(r.timeLayouts, layouts...)
		return nil
	}
}

// WithExtraTimeLayoutsWithNamedLocation configures
// the reader to use the given time layouts in
// addition to the built-in layouts when tyring to
// parse timestamps.
//
// These layouts can use names like `"EST"` to
// represent time zones. If the layouts do not use
// named locations, use [WithExtraTimeLayouts]
// instead.
func WithExtraTimeLayoutsWithNamedLocation(layouts ...string) Option {
	return func(r *Reader) error {
		// Check these use named layouts.
		for _, layout := range layouts {
			if !strings.Contains(layout, "MST") {
				return fmt.Errorf("cannot use layout %q with WithExtraTimeLayoutsWithNamedLocation: missing named layout", layout)
			}
		}

		r.timeLayoutsWithNamedLocation = append(r.timeLayoutsWithNamedLocation, layouts...)
		return nil
	}
}
