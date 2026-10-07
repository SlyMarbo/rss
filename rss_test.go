package rss

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"testing/iotest"
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
			t.Run("short reads", func(t *testing.T) {
				testar.Test(t, test, func(data io.Reader) (*Feed, error) {
					return parse(iotest.OneByteReader(data))
				}, "github.com/SlyMarbo/rss/v2")
			})
		})
	}
}

func TestParseRootElement(t *testing.T) {
	rss2 := `<rss version="2.0"><channel><title>RSS 2.0</title><item><guid>rss2-item</guid><title>RSS item</title></item></channel></rss>`
	rss1 := `<rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#" xmlns='http://purl.org/rss/1.0/'><channel><title>RSS 1.0</title></channel><item><guid>rss1-item</guid><title>RDF item</title></item></rdf:RDF>`
	atom := `<feed xmlns="http://www.w3.org/2005/Atom"><title>Atom</title><entry><id>atom-item</id><title>Atom item</title></entry></feed>`

	tests := []struct {
		name  string
		input string
		title string
		id    string
	}{
		{"RSS after a long comment", "<!--" + strings.Repeat("x", 1024) + "-->" + rss2, "RSS 2.0", "rss2-item"},
		{"RSS after long whitespace", strings.Repeat(" ", 1024) + rss2, "RSS 2.0", "rss2-item"},
		{"RSS with an RDF comment", `<!-- xmlns="http://purl.org/rss/1.0/" -->` + rss2, "RSS 2.0", "rss2-item"},
		{"RDF with single quotes", rss1, "RSS 1.0", "rss1-item"},
		{"RDF after a long comment", "<!--" + strings.Repeat("x", 1024) + "-->" + rss1, "RSS 1.0", "rss1-item"},
		{"RDF with an RSS comment", "<!-- <rss -->" + rss1, "RSS 1.0", "rss1-item"},
		{"Atom with an RSS comment", "<!-- <rss -->" + atom, "Atom", "atom-item"},
		{"Atom with an RDF comment", `<!-- xmlns="http://purl.org/rss/1.0/" -->` + atom, "Atom", "atom-item"},
		{"RSS with a prefixed root", strings.Replace(strings.Replace(rss2, `<rss version`, `<p:rss xmlns:p="urn:example:rss" version`, 1), "</rss>", "</p:rss>", 1), "RSS 2.0", "rss2-item"},
		{"RDF with a late namespace", strings.Replace(rss1, "xmlns=", `padding="`+strings.Repeat("x", 1024)+`" xmlns=`, 1), "RSS 1.0", "rss1-item"},
		{"RSS with a declared encoding", `<?xml version="1.0" encoding="ISO-8859-1"?>` + strings.Replace(rss2, "RSS 2.0", "caf\xe9", 1), "café", "rss2-item"},
	}

	r, err := NewReader()
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, short := range []bool{false, true} {
				t.Run(fmt.Sprintf("short=%t", short), func(t *testing.T) {
					var input io.Reader = strings.NewReader(test.input)
					if short {
						input = iotest.OneByteReader(input)
					}
					feed, err := r.Parse(defaultURL, input)
					if err != nil {
						t.Fatal(err)
					}
					if feed.Title != test.title || len(feed.Items) != 1 || feed.Items[0].ID != test.id {
						t.Fatalf("unexpected feed: %+v", feed)
					}
				})
			}
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
