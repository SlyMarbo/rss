package rss

import (
	"context"
	"io"
	"log/slog"

	"github.com/SlyMarbo/rss/v2/rss2"
)

func (r *Reader) parseRSS2(url string, data io.Reader) (*Feed, error) {
	rssFeed, err := rss2.Parse(data)
	if err != nil {
		return nil, err
	}

	return r.ConvertRSS2(url, rssFeed)
}

// ConvertRSS2 transforms data in RSS 1.0 format to
// this package's equivalent data structures.
//
// Most users should use [Reader.Parse] instead.
func (r *Reader) ConvertRSS2(url string, rssFeed *rss2.Feed) (*Feed, error) {
	ctx := context.Background()
	feed := &Feed{
		Title:       rssFeed.Channel.Title,
		Language:    rssFeed.Channel.Language,
		Description: rssFeed.Channel.Description,
		Items:       make([]*Item, 0, len(rssFeed.Channel.Items)),
		NextUpdate:  r.nextUpdate(rssFeed.Channel.MinsToLive, rssFeed.Channel.SkipHours, rssFeed.Channel.SkipDays),
		UpdateURL:   url,
	}

	if rssFeed.Channel.Author != "" {
		feed.Author = &Author{
			Name: rssFeed.Channel.Author,
		}
	}

	if rssFeed.Channel.Link != nil {
		feed.Links = make([]*Link, len(rssFeed.Channel.Link))
		for i, link := range rssFeed.Channel.Link {
			if link.Rel == "" && link.Type == "" && link.Href == "" && link.Chardata != "" {
				feed.Links[i] = &Link{
					Href: link.Chardata,
				}
			} else {
				feed.Links[i] = &Link{
					Href: link.Href,
					Rel:  link.Rel,
					Type: link.Type,
				}
			}
		}
	}

	if rssFeed.Channel.Image != nil {
		feed.Image = &Image{
			Title:  rssFeed.Channel.Image.Title,
			URL:    rssFeed.Channel.Image.URL,
			Height: uint32(rssFeed.Channel.Image.Height),
			Width:  uint32(rssFeed.Channel.Image.Width),
		}
	}

	if rssFeed.Channel.Categories != nil {
		feed.Categories = make([]string, len(rssFeed.Channel.Categories))
		for i, category := range rssFeed.Channel.Categories {
			feed.Categories[i] = category.Name
		}
	}

	// Process items.
	seen := make(map[string]struct{})
	for _, item := range rssFeed.Channel.Items {
		if item.ID == "" {
			if item.Link == "" {
				r.warn(ctx, "item has no ID or link and will be ignored", slog.String("item-title", item.Title))
				continue
			}

			item.ID = item.Link
		}

		// Skip items already seen.
		if _, ok := seen[item.ID]; ok {
			continue
		}

		next := &Item{
			Title:      item.Title,
			Summary:    item.Description,
			Content:    item.Content,
			Categories: item.Categories,
			ID:         item.ID,
			Read:       false,
		}

		if item.Link != "" {
			next.Links = []*Link{
				{Href: item.Link},
			}
		}

		if item.Date != "" {
			next.Date = r.parseTime(item.Date)
		} else if item.PubDate != "" {
			next.Date = r.parseTime(item.PubDate)
		}

		if len(item.Enclosures) > 0 {
			next.Enclosures = make([]*Enclosure, len(item.Enclosures))
			for i := range item.Enclosures {
				next.Enclosures[i] = &Enclosure{
					URL:    item.Enclosures[i].URL,
					Type:   item.Enclosures[i].Type,
					Length: item.Enclosures[i].Length,
				}
			}
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
