package rss

import (
	"fmt"
	"strings"
	"time"
)

// TimeLayoutsLoadLocation are time layouts
// which do not contain the location as a fixed
// constant. Instead of -0700, they use MST.
// Golang does not load the timezone by default,
// which means parseTime calls
// `time.LoadLocation(t.Location().String())`
// and then applies the offset returned by
// LoadLocation to the result.
var TimeLayoutsLoadLocation = []string{
	"Mon, 2 Jan 2006 15:04:05 MST",
	"Mon, 2 Jan 06 15:04:05 MST",
	"2 Jan 2006 15:04:05 MST",
	"2 Jan 06 15:04:05 MST",
	"Jan 2, 2006 15:04 PM MST",
	"Jan 2, 06 15:04 PM MST",

	time.RFC1123,
	time.RFC850,
	time.RFC822,
}

// TimeLayouts is contains a list of time.Parse() layouts that are used in
// attempts to convert item.Date and item.PubDate string to time.Time values.
// The layouts are attempted in ascending order until either time.Parse()
// does not return an error or all layouts are attempted.
var TimeLayouts = []string{
	"Mon, 2 Jan 2006 15:04:05 Z",
	"Mon, 2 Jan 2006 15:04:05",
	"Mon, 2 Jan 2006 15:04:05 -0700",
	"Mon, 2 Jan 06 15:04:05 -0700",
	"Mon, 2 Jan 06 15:04:05",
	"2 Jan 2006 15:04:05 -0700",
	"2 Jan 2006 15:04:05",
	"2 Jan 06 15:04:05 -0700",
	"2006-01-02 15:04:05 -0700",
	"2006-01-02 15:04:05",
	time.ANSIC,
	time.UnixDate,
	time.RubyDate,
	time.RFC822Z,
	time.RFC1123Z,
	time.RFC3339,
	time.RFC3339Nano,

	// Not some common time zones. While they are odd,
	// they can occur but we should only check
	// them last (to slightly improve runtime).
	// We also always use the offset, because the
	// texual representation can be ambiguous.
	// For example, PST can have different rules
	// in different locals:
	// https://news.ycombinator.com/item?id=10199812
	"2 Jan 2006 15:04:05 -0700 MST",
	"2 Jan 2006 15:04:05 MST -0700",
	"Mon, 2 Jan 2006 15:04:05 MST -0700",
	"Mon, 2 Jan 2006 15:04:05 -0700 MST",
	"2 Jan 06 15:04:05 -0700 MST",
	"2 Jan 06 15:04:05 MST -0700",
	"Jan 2, 2006 15:04 PM -0700 MST",
	"Jan 2, 2006 15:04 PM MST -0700",
	"Jan 2, 06 15:04 PM MST -0700",
	"Jan 2, 06 15:04 PM -0700 MST",
}

func parseTime(s string) (time.Time, error) {
	s = strings.TrimSpace(s)

	var e error
	var t time.Time

	for _, layout := range TimeLayouts {
		t, e = time.Parse(layout, s)
		if e == nil {
			return t, nil
		}
	}

	for _, layout := range TimeLayoutsLoadLocation {
		t, e = time.Parse(layout, s)
		if e != nil {
			continue
		}

		// In case LoadLocation returns an error
		// we want to return the time and error
		// as is. LoadLocation commonly returns an
		// error if tzinfo is not installed.
		// This often happens when running go applications
		// inside an alpine docker container.
		loc, err := time.LoadLocation(t.Location().String())
		if err != nil {
			if debug {
				fmt.Printf("[w] could not load timezone: %q\n", e)
			}
			return t, err
		}

		t, e = time.ParseInLocation(layout, s, loc)
		if e == nil {
			return t, nil
		}
	}

	return time.Time{}, e
}
