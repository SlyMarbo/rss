package rss

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"text/tabwriter"
	"time"
)

// Parse RSS or Atom data.
func Parse(data []byte) (*Feed, error) {

	if strings.Contains(string(data), "<rss") {
		if debug {
			fmt.Println("[i] Parsing as RSS 2.0")
		}
		return parseRSS2(data, database)
	} else if strings.Contains(string(data), "xmlns=\"http://purl.org/rss/1.0/\"") {
		if debug {
			fmt.Println("[i] Parsing as RSS 1.0")
		}
		return parseRSS1(data, database)
	} else {
		if debug {
			fmt.Println("[i] Parsing as Atom")
		}
		return parseAtom(data, database)
	}
}

// CacheParsedItemIDs enables or disable Item.ID caching when parsing feeds.
// Returns whether Item.ID were cached prior to function call.
func CacheParsedItemIDs(flag bool) (didCache bool) {
	didCache = !disabled
	disabled = !flag
	return
}

type FetchFunc func() (resp *http.Response, err error)

// Fetch downloads and parses the RSS feed at the given URL
func Fetch(url string) (*Feed, error) {
	return FetchByClient(url, http.DefaultClient)
}

func FetchByClient(url string, client *http.Client) (*Feed, error) {
	fetchFunc := func() (resp *http.Response, err error) {
		return client.Get(url)
	}
	return FetchByFunc(fetchFunc, url)
}

func FetchByFunc(fetchFunc FetchFunc, url string) (*Feed, error) {
	resp, err := fetchFunc()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	out, err := Parse(body)
	if err != nil {
		return nil, err
	}

	if out.Link == "" {
		out.Link = url
	}

	out.UpdateURL = url

	return out, nil
}

// Feed is the top-level structure.
type Feed struct {
	Nickname    string              `json:"nickname"` // This is not set by the package, but could be helpful.
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Link        string              `json:"link"`      // Link to the creator's website.
	UpdateURL   string              `json:"updateurl"` // URL of the feed itself.
	Image       *Image              `json:"image"`     // Feed icon.
	Items       []*Item             `json:"items"`
	ItemMap     map[string]struct{} `json:"itemmap"` // Used in checking whether an item has been seen before.
	Refresh     time.Time           `json:"refresh"` // Earliest time this feed should next be checked.
	Unread      uint32              `json:"unread"`  // Number of unread items. Used by aggregators.
}

// Update fetches any new items and updates f.
func (f *Feed) Update() error {

	// Check that we don't update too often.
	if f.Refresh.After(time.Now()) {
		return nil
	}

	if f.UpdateURL == "" {
		return errors.New("Error: feed has no URL.")
	}

	if f.ItemMap == nil {
		f.ItemMap = make(map[string]struct{})
		for _, item := range f.Items {
			if _, ok := f.ItemMap[item.ID]; !ok {
				f.ItemMap[item.ID] = struct{}{}
			}
		}
	}

	update, err := Fetch(f.UpdateURL)
	if err != nil {
		return err
	}

	f.Refresh = update.Refresh
	f.Title = update.Title
	f.Description = update.Description

	for _, item := range update.Items {
		if _, ok := f.ItemMap[item.ID]; !ok {
			f.Items = append(f.Items, item)
			f.ItemMap[item.ID] = struct{}{}
			f.Unread++
		}
	}

	return nil
}

func (f *Feed) String() string {
	buf := new(bytes.Buffer)
	if debug {
		w := tabwriter.NewWriter(buf, 0, 8, 0, '\t', tabwriter.StripEscape)
		fmt.Fprintf(w, "Feed {\n")
		fmt.Fprintf(w, "\xff\t\xffNickname:\t%q\n", f.Nickname)
		fmt.Fprintf(w, "\xff\t\xffTitle:\t%q\n", f.Title)
		fmt.Fprintf(w, "\xff\t\xffDescription:\t%q\n", f.Description)
		fmt.Fprintf(w, "\xff\t\xffLink:\t%q\n", f.Link)
		fmt.Fprintf(w, "\xff\t\xffUpdateURL:\t%q\n", f.UpdateURL)
		fmt.Fprintf(w, "\xff\t\xffImage:\t%q (%s)\n", f.Image.Title, f.Image.Url)
		fmt.Fprintf(w, "\xff\t\xffRefresh:\t%s\n", f.Refresh.Format(DATE))
		fmt.Fprintf(w, "\xff\t\xffUnread:\t%d\n", f.Unread)
		fmt.Fprintf(w, "\xff\t\xffItems:\t(%d) {\n", len(f.Items))
		for _, item := range f.Items {
			fmt.Fprintf(w, "%s\n", item.Format(2))
		}
		fmt.Fprintf(w, "\xff\t\xff}\n}\n")
		w.Flush()
	} else {
		w := buf
		fmt.Fprintf(w, "Feed %q\n", f.Title)
		fmt.Fprintf(w, "\t%q\n", f.Description)
		fmt.Fprintf(w, "\t%q\n", f.Link)
		fmt.Fprintf(w, "\t%s\n", f.Image)
		fmt.Fprintf(w, "\tRefresh at %s\n", f.Refresh.Format(DATE))
		fmt.Fprintf(w, "\tUnread: %d\n", f.Unread)
		fmt.Fprintf(w, "\tItems:\n")
		for _, item := range f.Items {
			fmt.Fprintf(w, "\t%s\n", item.Format(2))
		}
	}
	return buf.String()
}

// Item represents a single story.
type Item struct {
	Title      string       `json:"title"`
	Summary    string       `json:"summary"`
	Content    string       `json:"content"`
	Link       string       `json:"link"`
	Date       time.Time    `json:"date"`
	ID         string       `json:"id"`
	Enclosures []*Enclosure `json:"enclosures"`
	Read       bool         `json:"read"`
}

func (i *Item) String() string {
	return i.Format(0)
}

func (i *Item) Format(indent int) string {
	buf := new(bytes.Buffer)
	single := strings.Repeat("\t", indent)
	double := single + "\t"
	if debug {
		w := tabwriter.NewWriter(buf, 0, 8, 0, '\t', tabwriter.StripEscape)
		fmt.Fprintf(w, "\xff%s\xffItem {\n", single)
		fmt.Fprintf(w, "\xff%s\xffTitle:\t%q\n", double, i.Title)
		fmt.Fprintf(w, "\xff%s\xffSummary:\t%q\n", double, i.Summary)
		fmt.Fprintf(w, "\xff%s\xffLink:\t%s\n", double, i.Link)
		fmt.Fprintf(w, "\xff%s\xffDate:\t%s\n", double, i.Date.Format(DATE))
		fmt.Fprintf(w, "\xff%s\xffID:\t%s\n", double, i.ID)
		fmt.Fprintf(w, "\xff%s\xffRead:\t%v\n", double, i.Read)
		fmt.Fprintf(w, "\xff%s\xffContent:\t%q\n", double, i.Content)
		fmt.Fprintf(w, "\xff%s\xff}\n", single)
		w.Flush()
	} else {
		w := buf
		fmt.Fprintf(w, "%sItem %q\n", single, i.Title)
		fmt.Fprintf(w, "%s%q\n", double, i.Link)
		fmt.Fprintf(w, "%s%s\n", double, i.Date.Format(DATE))
		fmt.Fprintf(w, "%s%q\n", double, i.ID)
		fmt.Fprintf(w, "%sRead: %v\n", double, i.Read)
		fmt.Fprintf(w, "%s%q\n", double, i.Content)
	}
	return buf.String()
}

type Enclosure struct {
	Url    string `json:"url"`
	Type   string `json:"type"`
	Rel    string `json:"rel"`
	Length int    `json:"length"`
}

func (e *Enclosure) Get() (io.ReadCloser, error) {
	if e == nil || e.Url == "" {
		return nil, errors.New("No enclosure")
	}

	res, err := http.Get(e.Url)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

type Image struct {
	Title  string `json:"title"`
	Url    string `json:"url"`
	Height uint32 `json:"height"`
	Width  uint32 `json:"width"`
}

func (i *Image) Get() (io.ReadCloser, error) {
	if i == nil || i.Url == "" {
		return nil, errors.New("No image")
	}

	res, err := http.Get(i.Url)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

func (i *Image) String() string {
	return fmt.Sprintf("Image %q", i.Title)
}
