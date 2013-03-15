package rss

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var database *db

func init() {
	database = NewDB()
	go database.Run()
}

func Parse(data []byte) (*Feed, error) {
	if strings.Contains(string(data), "<rss") {
		return parseRSS2(data, database)
	} else if strings.Contains(string(data), "xmlns=\"http://purl.org/rss/1.0/\"") {
		return parseRSS1(data, database)
	} else {
		return parseAtom(data, database)
	}
	
	panic("Unreachable.")
}

func parseRSS2(data []byte, read *db) (*Feed, error) {
	feed := rss2_0Feed{}
	p := xml.NewDecoder(bytes.NewReader(data))
	p.CharsetReader = charsetReader
	err := p.Decode(&feed)
	if err != nil {
		return nil, err
	}
	if feed.Channel == nil {
		return nil, fmt.Errorf("Error: no channel found in %q.", string(data))
	}
	
	channel := feed.Channel
	
	out := new(Feed)
	out.Title = channel.Title
	out.Description = channel.Description
	out.Link = channel.Link
	out.Image = channel.Image.Image()
	if channel.MinsToLive != 0 {
		sort.Ints(channel.SkipHours)
		next := time.Now().Add(time.Duration(channel.MinsToLive) * time.Minute)
		for _, hour := range channel.SkipHours {
			if hour == next.Hour() {
				next.Add(time.Duration(60 - next.Minute()) * time.Minute)
			}
		}
		trying := true
		for trying {
			trying = false
			for _, day := range channel.SkipDays {
				if strings.Title(day) == next.Weekday().String() {
					next.Add(time.Duration(24 - next.Hour()) * time.Hour)
					trying = true
					break
				}
			}
		}
		
		out.Refresh = next
	}
	
	if out.Refresh.IsZero() {
		out.Refresh = time.Now().Add(10 * time.Minute)
	}
	
	if channel.Items == nil {
		return nil, fmt.Errorf("Error: no feeds found in %q.", string(data))
	}
	
	out.Items = make([]*Item, 0, len(channel.Items))
	
	// Process items.
	for _, item := range channel.Items {
		
		// Skip items already known.
		if read.req <- item.ID; <- read.res {
			continue
		}
		
		next := new(Item)
		next.Title = item.Title
		next.Content = item.Content
		next.Link = item.Link
		if item.Date != "" {
			next.Date, err = parseTime(item.Date)
			if err != nil {
				return nil, err
			}
		}
		next.ID = item.ID
		next.Read = false
		
		out.Items = append(out.Items, next)
		out.Unread++
	}
	
	return out, nil
}

func parseTime(s string) (time.Time, error) {
	formats := []string{
		"Mon, _2 Jan 2006 15:04:05 MST",
		"Mon, _2 Jan 2006 15:04:05 -0700",
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.RFC822,
		time.RFC822Z,
		time.RFC850,
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
		time.RFC3339Nano,
	}
	
	var e error
	var t time.Time
	
	for _, format := range formats {
		t, e = time.Parse(format, s)
		if e == nil {
			return t, e
		}
	}
	
	return time.Time{}, e
}

// External structures.

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

func (f *Feed) String() string {
	buf := new(bytes.Buffer)
	buf.WriteString(fmt.Sprintf("Feed %q\n\t%q\n\t%q\n\t%s\n\tRefresh at %s\n\tUnread: %d\n\tItems:\n",
		f.Title, f.Description, f.Link, f.Image, f.Refresh.Format("Mon 2 Jan 2006 15:04:05 MST"), f.Unread))
	for _, item := range f.Items {
		buf.WriteString(fmt.Sprintf("\t%s\n", item.Format("\t\t")))
	}
	return buf.String()
}

func (i *Item) String() string {
	return i.Format("")
}

func (i *Item) Format(s string) string {
	return fmt.Sprintf("Item %q\n\t%s%q\n\t%s%s\n\t%s%q\n\t%sRead: %v\n\t%s%q", i.Title, s, i.Link, s,
		i.Date.Format("Mon 2 Jan 2006 15:04:05 MST"), s, i.ID, s, i.Read, s, i.Content)
}

func (i *Image) String() string {
	return fmt.Sprintf("Image %q", i.Title)
}

// Internal structures.

type db struct {
	req chan string
	res chan bool
	known map[string]struct{}
}

func (d *db) Run() {
	d.known = make(map[string]struct{})
	var s string
	
	for {
		s = <- d.req
		if _, ok := d.known[s]; ok {
			d.res <- true
		} else {
			d.res <- false
			d.known[s] = struct{}{}
		}
	}
}

func NewDB() *db {
	out := new(db)
	out.req = make(chan string)
	out.res = make(chan bool)
	return out
}

// RSS

type rss2_0Feed struct {
	XMLName  xml.Name       `xml:"rss"`
	Channel  *rss2_0Channel `xml:"channel"`
}

type rss2_0Channel struct {
	XMLName     xml.Name     `xml:"channel"`
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	Link        string       `xml:"link"`
	Image       rss2_0Image  `xml:"image"`
	Items       []rss2_0Item `xml:"item"`
	MinsToLive  int          `xml:"ttl"`
	SkipHours   []int        `xml:"skipHours>hour"`
	SkipDays    []string     `xml:"skipDays>day"`
}

type rss2_0Item struct {
	XMLName xml.Name `xml:"item"`
	Title   string   `xml:"title"`
	Content string   `xml:"description"`
	Link    string   `xml:"link"`
	Date    string   `xml:"pubDate"`
	ID      string   `xml:"guid"`
}

type rss2_0Image struct {
	XMLName xml.Name `xml:"image"`
	Title   string   `xml:"title"`
	Url     string   `xml:"url"`
	Height  int      `xml:"height"`
	Width   int      `xml:"width"`
}

func (i *rss2_0Image) Image() *Image {
	out := new(Image)
	out.Title = i.Title
	out.Url = i.Url
	out.Height = uint32(i.Height)
	out.Width = uint32(i.Width)
	return out
}

// ISO-8859-1 support

type charsetISO88591er struct {
    r   io.ByteReader
    buf *bytes.Buffer
}

func newCharsetISO88591(r io.Reader) *charsetISO88591er {
    buf := bytes.NewBuffer(make([]byte, 0, utf8.UTFMax))
    return &charsetISO88591er{r.(io.ByteReader), buf}
}

func (cs *charsetISO88591er) ReadByte() (b byte, err error) {
    // http://unicode.org/Public/MAPPINGS/ISO8859/8859-1.TXT
    // Date: 1999 July 27; Last modified: 27-Feb-2001 05:08
    if cs.buf.Len() <= 0 {
        r, err := cs.r.ReadByte()
        if err != nil {
            return 0, err
        }
        if r < utf8.RuneSelf {
            return r, nil
        }
        cs.buf.WriteRune(rune(r))
    }
    return cs.buf.ReadByte()
}

func (cs *charsetISO88591er) Read(p []byte) (int, error) {
    // Use ReadByte method.
    return 0, errors.New("Use ReadByte()")
}

func isCharset(charset string, names []string) bool {
    charset = strings.ToLower(charset)
    for _, n := range names {
        if charset == strings.ToLower(n) {
            return true
        }
    }
    return false
}

func isCharsetISO88591(charset string) bool {
    // http://www.iana.org/assignments/character-sets
    // (last updated 2010-11-04)
    names := []string{
        // Name
        "ISO_8859-1:1987",
        // Alias (preferred MIME name)
        "ISO-8859-1",
        // Aliases
        "iso-ir-100",
        "ISO_8859-1",
        "latin1",
        "l1",
        "IBM819",
        "CP819",
        "csISOLatin1",
    }
    return isCharset(charset, names)
}

func isCharsetUTF8(charset string) bool {
    names := []string{
        "UTF-8",
        // Default
        "",
    }
    return isCharset(charset, names)
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
    switch {
    case isCharsetUTF8(charset):
        return input, nil
    case isCharsetISO88591(charset):
        return newCharsetISO88591(input), nil
    }
    return nil, errors.New("CharsetReader: unexpected charset: " + charset)
}
