package rss

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/synctest"
	"time"

	"github.com/SlyMarbo/rss/v2/internal/testar"

	"golang.org/x/tools/txtar"
)

const defaultURL = "https://example.com/feed.xml"

func TestParse(t *testing.T) {
	tests := []string{
		"atom-1.0-1",
		"atom-1.0-2",
		"atom-1.0-enclosure-1",
		"atom-1.0-html-1",
		"feedparser",
		"google-merchant-center",
		"rss-0.91-1",
		"rss-0.92-1",
		"rss-1.0-1",
		"rss-1.0-enclosure-1",
		"rss-1.0-no-channel-1",
		"rss-2.0-1",
		"rss-2.0-2",
		"rss-2.0-content-encoded-1",
		"rss-2.0-enclosure-1",
		"rss-2.0-enclosure-2",
		"rss-2.0-no-channel-1",
		"rss-2.0-update-1",
		"rss-2.0-update-2",
		"sample-rss-1-1",
		"sample-rss-1-modules-1",
		"sample-rss-2-1",
	}

	now := func() time.Time {
		return time.Date(2025, time.November, 22, 12, 13, 14, 15, time.UTC)
	}

	r, err := NewReader(WithNow(now))
	if err != nil {
		t.Fatal(err)
	}

	parse := func(data io.Reader) (*Feed, error) {
		return r.Parse(defaultURL, data)
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			testar.Test(t, test, parse, "github.com/SlyMarbo/rss/v2")
		})
	}
}

func TestFetch(t *testing.T) {
	tests := []string{
		"atom-1.0-1",
		"sample-rss-1-1",
	}

	ctx := context.Background()

	now := func() time.Time {
		return time.Date(2025, time.November, 22, 12, 13, 14, 15, time.UTC)
	}

	mux := http.NewServeMux()
	archives := make(map[string]*txtar.Archive, len(tests))
	for _, test := range tests {
		ar, err := txtar.ParseFile(filepath.Join("testdata", test) + ".txtar")
		if err != nil {
			t.Fatalf("failed to parse test vector %q: %v", test, err)
		}

		archives[test] = ar
		mux.HandleFunc("GET /"+test, func(w http.ResponseWriter, r *http.Request) {
			http.ServeContent(w, r, test+".xml", time.Time{}, bytes.NewReader(ar.Files[2].Data))
		})
	}

	// Add a default route.
	mux.Handle("/", http.NotFoundHandler())

	srv := httptest.NewTLSServer(mux)
	defer srv.Close()

	client := srv.Client()

	r, err := NewReader(
		WithNow(now),
		WithHTTPClient(client),
	)
	if err != nil {
		t.Fatal(err)
	}

	// We can be more lax in handling the txtar
	// data, as it's validated very strictly in
	// testar, which is used above.

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			ar := archives[test]
			url := srv.URL + "/" + test
			feed, err := r.Fetch(ctx, url)

			// Fix up the URLs.
			if feed != nil {
				if len(feed.Links) == 1 && feed.Links[0].Href == url {
					feed.Links[0].Href = defaultURL
				}

				if feed.UpdateURL == url {
					feed.UpdateURL = defaultURL
				}
			}

			testar.Check(t, ar, feed, err)
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []string{
		"sample-rss-1-1",
	}

	ctx := context.Background()

	epoch := time.Date(2025, time.November, 22, 12, 13, 14, 15, time.UTC)
	now := func() time.Time {
		return epoch
	}

	mux := http.NewServeMux()
	archives := make(map[string]*txtar.Archive, len(tests))
	for _, test := range tests {
		ar, err := txtar.ParseFile(filepath.Join("testdata", test) + ".txtar")
		if err != nil {
			t.Fatalf("failed to parse test vector %q: %v", test, err)
		}

		archives[test] = ar
		mux.HandleFunc("GET /"+test, func(w http.ResponseWriter, r *http.Request) {
			http.ServeContent(w, r, test+".xml", time.Time{}, bytes.NewReader(ar.Files[2].Data))
		})
	}

	// Add a default route.
	mux.Handle("/", http.NotFoundHandler())

	srv := httptest.NewServer(mux) // Don't use TLS, as our time weirdness breaks validity periods.
	defer srv.Close()

	client := srv.Client()

	r, err := NewReader(
		WithNow(now),
		WithHTTPClient(client),
	)
	if err != nil {
		t.Fatal(err)
	}

	// We can be more lax in handling the txtar
	// data, as it's validated very strictly in
	// testar, which is used above.

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			ar := archives[test]
			url := srv.URL + "/" + test
			feed, err := r.Parse(url, bytes.NewReader(ar.Files[2].Data))
			if err != nil {
				t.Fatalf("failed to parse test vector: %v", err)
			}

			t.Run("early", func(t *testing.T) {
				// Set NextUpdate to an hour in the
				// future and check Update returns
				// ErrTooSoon. Drop the last item so
				// we can see whether an update has
				// taken effect.
				feed.UpdateURL = url
				feed.NextUpdate = epoch.Add(time.Hour)
				feed.Items = feed.Items[:len(feed.Items)-1]
				if feed.Unread > 0 {
					feed.Unread--
				}

				err := r.Update(ctx, feed)
				if !errors.Is(err, ErrTooSoon) {
					t.Fatalf("got unexpected error when updating too early:\nGot:  %v\nWant: %v", err, ErrTooSoon)
				}

				// Now actually wait.
				synctest.Test(t, func(t *testing.T) {
					err = r.waitForFeed(ctx, feed)
					if err != nil {
						t.Fatal(err)
					}
				})

				err = r.update(ctx, feed)

				// Fix up the URLs.
				if feed != nil {
					feed.NextUpdate = epoch.Add(r.defaultUpdateInterval)
					if len(feed.Links) == 1 && feed.Links[0].Href == url {
						feed.Links[0].Href = defaultURL
					}

					if feed.UpdateURL == url {
						feed.UpdateURL = defaultURL
					}
				}

				// Check the data.
				testar.Check(t, ar, feed, nil)
			})

			t.Run("ready", func(t *testing.T) {
				// Set NextUpdate to an hour in the
				// past and check Update works now.
				// Drop the last item so we can see
				// we can see whether an update has
				// taken effect.
				feed.UpdateURL = url
				feed.NextUpdate = epoch.Add(-time.Hour)
				feed.Items = feed.Items[:len(feed.Items)-1]
				if feed.Unread > 0 {
					feed.Unread--
				}

				err := r.Update(ctx, feed)
				if err != nil {
					t.Fatalf("got unexpected error when updating on time: %v", err)
				}

				// Fix up the URLs.
				if feed != nil {
					feed.NextUpdate = epoch.Add(r.defaultUpdateInterval)
					if len(feed.Links) == 1 && feed.Links[0].Href == url {
						feed.Links[0].Href = defaultURL
					}

					if feed.UpdateURL == url {
						feed.UpdateURL = defaultURL
					}
				}

				// Check the data.
				testar.Check(t, ar, feed, nil)
			})

			t.Run("ready-patiently", func(t *testing.T) {
				// Set NextUpdate to an hour in the
				// past and check Update works now.
				// Drop the last item so we can see
				// we can see whether an update has
				// taken effect.
				feed.UpdateURL = url
				feed.NextUpdate = epoch.Add(-time.Hour)
				feed.Items = feed.Items[:len(feed.Items)-1]
				if feed.Unread > 0 {
					feed.Unread--
				}

				err := r.UpdatePatiently(ctx, feed)
				if err != nil {
					t.Fatalf("got unexpected error when updating on time: %v", err)
				}

				// Fix up the URLs.
				if feed != nil {
					feed.NextUpdate = epoch.Add(r.defaultUpdateInterval)
					if len(feed.Links) == 1 && feed.Links[0].Href == url {
						feed.Links[0].Href = defaultURL
					}

					if feed.UpdateURL == url {
						feed.UpdateURL = defaultURL
					}
				}

				// Check the data.
				testar.Check(t, ar, feed, nil)
			})
		})
	}
}
