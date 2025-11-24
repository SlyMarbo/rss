package rss

import (
	"context"
	"io"
	"log/slog"

	"github.com/SlyMarbo/rss/v2/atom1"
)

func (r *Reader) parseAtom1(url string, data io.Reader) (*Feed, error) {
	atomFeed, err := atom1.Parse(data)
	if err != nil {
		return nil, err
	}

	return r.ConvertAtom1(url, atomFeed)
}

// ConvertAtom1 transforms data in Atom 1.0 format to
// this package's equivalent data structures.
//
// Most users should use [Reader.Parse] instead.
func (r *Reader) ConvertAtom1(url string, atomFeed *atom1.Feed) (*Feed, error) {
	ctx := context.Background()
	feed := &Feed{
		Title:       atomFeed.Title,
		Description: atomFeed.Description,
		NextUpdate:  r.now().Add(r.defaultUpdateInterval),
		Items:       make([]*Item, 0, len(atomFeed.Items)),
		UpdateURL:   url,
	}

	if atomFeed.Image != nil {
		feed.Image = &Image{
			Title:  atomFeed.Image.Title,
			URL:    atomFeed.Image.URL,
			Height: uint32(atomFeed.Image.Height),
			Width:  uint32(atomFeed.Image.Width),
		}
	}

	if atomFeed.Author != nil && (atomFeed.Author.Name != "" || atomFeed.Author.URI != "" || atomFeed.Author.Email != "") {
		feed.Author = &Author{
			Name:  atomFeed.Author.Name,
			URL:   atomFeed.Author.URI,
			Email: atomFeed.Author.Email,
		}

		if atomFeed.Author.Links != nil {
			feed.Author.Links = make([]*Link, len(atomFeed.Author.Links))
			for i, link := range atomFeed.Author.Links {
				feed.Author.Links[i] = &Link{
					Href: link.Href,
					Rel:  link.Rel,
					Type: link.Type,
				}
			}
		}
	}

	if atomFeed.Link != nil {
		feed.Links = make([]*Link, len(atomFeed.Link))
		for i, link := range atomFeed.Link {
			feed.Links[i] = &Link{
				Href: link.Href,
				Rel:  link.Rel,
				Type: link.Type,
			}
		}
	}

	// Process items.
	seen := make(map[string]struct{})
	for _, item := range atomFeed.Items {
		// Skip items already seen.
		if _, ok := seen[item.ID]; ok {
			continue
		}

		if item.ID == "" {
			r.warn(ctx, "item has no ID or link and will be ignored", slog.String("item-title", item.Title))
			continue
		}

		next := &Item{
			Title:   item.Title,
			Summary: item.Summary,
			ID:      item.ID,
			Read:    false,
		}

		if item.Content != nil {
			next.Content = item.Content.RAWContent
		}

		if item.Date != "" {
			next.Date = r.parseTime(item.Date)
		}

		for _, link := range item.Links {
			next.Links = append(next.Links, &Link{
				Href: link.Href,
				Rel:  link.Rel,
				Type: link.Type,
			})
		}

		seen[next.ID] = struct{}{}
		feed.Items = append(feed.Items, next)
		feed.Unread++
	}

	if feed.Links == nil {
		feed.Links = []*Link{
			{Href: url},
		}
	}

	return feed, nil
}
