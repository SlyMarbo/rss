package rss

import (
	"slices"
	"strings"
	"time"
)

// nextUpdate converts RSS-style update data to the
// timestamp of the next acceptable update.
func (r *Reader) nextUpdate(minsToLive int, skipHours []int, skipDays []string) time.Time {
	return nextUpdate(r.now(), r.defaultUpdateInterval, minsToLive, skipHours, skipDays)
}

func nextUpdate(now time.Time, defaultInterval time.Duration, minsToLive int, skipHours []int, skipDays []string) time.Time {
	if minsToLive == 0 {
		return now.Add(defaultInterval)
	}

	skipHours = slices.Clone(skipHours)
	slices.Sort(skipHours)
	update := now.Add(time.Duration(minsToLive) * time.Minute)

	for _, hour := range skipHours {
		if hour == update.Hour() {
			update = update.Add(time.Hour)
		}
	}

again:
	for _, day := range skipDays {
		if strings.Title(day) == update.Weekday().String() {
			update = update.Add(time.Duration(24-update.Hour()) * time.Hour)
			goto again
		}
	}

	return update
}

func (r *Reader) parseTime(s string) time.Time {
	return parseTime(s, r.timeLayouts, r.timeLayoutsWithNamedLocation)
}

func parseTime(s string, layouts, layoutsWithNamedLocation []string) time.Time {
	s = strings.TrimSpace(s)
	for _, layout := range layouts {
		tv, err := time.Parse(layout, s)
		if err == nil {
			return tv.UTC()
		}
	}

	for _, layout := range layoutsWithNamedLocation {
		tv, err := time.Parse(layout, s)
		if err != nil {
			continue
		}

		loc, err := time.LoadLocation(tv.Location().String())
		if err != nil {
			return time.Time{}
		}

		tv, err = time.ParseInLocation(layout, s, loc)
		if err == nil {
			return tv.UTC()
		}
	}

	return time.Time{}
}

// defaultTimeLayouts contains a list of [time.Parse]
// layouts that are used in attempts to convert
// item.Date and item.PubDate strings to time.Time
// values.
//
// The layouts are attempted sequentially until
// either [time.Parse] does not return an error or
// all layouts have been attempted.
var defaultTimeLayouts = []string{
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

// defaultTimeLayoutsWithNamedLocation are time
// layouts that do not contain the location as
// a literal constant. For example, instead of
// -0700, they use MST.
//
// Go does not load the timezone by default,
// which means [parseTime] calls
// `time.LoadLocation(t.Location().String())`
// and then applies the offset returned by
// LoadLocation to the result.
var defaultTimeLayoutsWithNamedLocation = []string{
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
