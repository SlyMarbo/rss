package rss

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

func parseJsonFeedV1(data []byte) (*Feed, error) {
	warnings := false
	feed := json_v1Feed{}
	p := json.NewDecoder(bytes.NewReader(data))
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}

	out := new(Feed)
	out.Title = feed.Title
	out.Description = feed.Description
	out.Link = feed.HomePageURL
	out.UpdateURL = feed.FeedURL
	out.Image = &Image{URL: feed.Favicon}
	out.Refresh = time.Now().Add(10 * time.Minute)

	out.Items = make([]*Item, 0, len(feed.Items))
	out.ItemMap = make(map[string]struct{})

	// Process items.
	for _, item := range feed.Items {

		// Skip items already known.
		if _, ok := out.ItemMap[item.ID]; ok {
			continue
		}

		next := new(Item)
		next.Title = item.Title
		next.Summary = item.Summary
		if item.ContentHTML != "" {
			next.Content = item.ContentHTML
		} else {
			next.Content = item.ContentText
		}

		date := item.DateModified
		if date == "" {
			date = item.DatePublished
		}
		if date != "" {
			next.Date, err = parseTime(date)
			if err == nil {
				next.DateValid = true
			}
		}
		next.ID = item.ID
		next.Link = item.URL
		for _, attachment := range item.Attachments {
			next.Enclosures = append(next.Enclosures, &Enclosure{
				URL:    attachment.URL,
				Type:   attachment.MIMEType,
				Length: attachment.DurationInSeconds,
			})
		}
		next.Read = false

		if next.ID == "" {
			if debug {
				fmt.Printf("[w] Item %q has no ID and will be ignored.\n", next.Title)
				fmt.Printf("[w] %#v\n", item)
			}
			warnings = true
			continue
		}

		if _, ok := out.ItemMap[next.ID]; ok {
			if debug {
				fmt.Printf("[w] Item %q has duplicate ID.\n", next.Title)
				fmt.Printf("[w] %#v\n", next)
			}
			warnings = true
			continue
		}

		out.Items = append(out.Items, next)
		out.ItemMap[next.ID] = struct{}{}
		out.Unread++
	}

	if warnings && debug {
		fmt.Printf("[i] Encountered warnings:\n%s\n", data)
	}

	return out, nil
}

type json_v1Item struct {
	ID            string              `json:"id"`
	URL           string              `json:"url"`
	ExternalURL   string              `json:"external_url"`
	Title         string              `json:"title"`
	ContentHTML   string              `json:"content_html"`
	ContentText   string              `json:"content_text"`
	Summary       string              `json:"summary"`
	Image         string              `json:"image"`
	BannerImage   string              `json:"banner_image"`
	DatePublished string              `json:"date_published"`
	DateModified  string              `json:"date_modified"`
	Author        json_v1Author       `json:"author"`
	Tags          []string            `json:"tags"`
	Attachments   []json_v1Attachment `json:"attachments"`
}

type json_v1Author struct {
	Name   string `json:"name"`
	URL    string `json:"url"`
	Avatar string `json:"avatar"`
}

type json_v1Hub struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

type json_v1Attachment struct {
	URL               string `json:"url"`
	MIMEType          string `json:"mime_type"`
	Title             string `json:"title"`
	SizeInBytes       uint   `json:"size_in_bytes"`
	DurationInSeconds uint   `json:"duration_in_seconds"`
}

type json_v1Feed struct {
	Version     string        `json:"version"`
	Title       string        `json:"title"`
	HomePageURL string        `json:"home_page_url"`
	FeedURL     string        `json:"feed_url"`
	Description string        `json:"Description"`
	UserComment string        `json:"user_comment"`
	NextURL     string        `json:"next_url"`
	Icon        string        `json:"icon"`
	Favicon     string        `json:"favicon"`
	Author      json_v1Author `json:"author"`
	Expired     bool          `json:"expired"`
	Hubs        []json_v1Hub  `json:"hubs"`
	Items       []json_v1Item `json:"items"`
}
