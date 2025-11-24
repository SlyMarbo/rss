package rss

import (
	"testing"
	"time"
)

func timestamp(year int, month time.Month, day, hour, minute int) time.Time {
	return time.Date(year, month, day, hour, minute, 0, 0, time.UTC)
}

func TestNextUpdate(t *testing.T) {
	tests := []struct {
		Name            string
		Now             time.Time
		DefaultInterval time.Duration
		MinsToLive      int
		SkipHours       []int
		SkipDays        []string
		Want            time.Time
	}{
		{
			Name:            "default",
			Now:             timestamp(2006, time.January, 2, 15, 4),
			DefaultInterval: time.Hour,
			Want:            timestamp(2006, time.January, 2, 16, 4),
		},
		{
			Name:       "ttl only",
			Now:        timestamp(2006, time.January, 2, 15, 4),
			MinsToLive: 5,
			Want:       timestamp(2006, time.January, 2, 15, 9),
		},
		{
			Name:       "ttl ignored hour skip",
			Now:        timestamp(2006, time.January, 2, 15, 4),
			MinsToLive: 5,
			SkipHours:  []int{16},
			Want:       timestamp(2006, time.January, 2, 15, 9),
		},
		{
			Name:       "ttl skip one hour",
			Now:        timestamp(2006, time.January, 2, 15, 4),
			MinsToLive: 5,
			SkipHours:  []int{15},
			Want:       timestamp(2006, time.January, 2, 16, 9),
		},
		{
			Name:       "ttl skip two hours",
			Now:        timestamp(2006, time.January, 2, 15, 4),
			MinsToLive: 5,
			SkipHours:  []int{15, 16, 18},
			Want:       timestamp(2006, time.January, 2, 17, 9),
		},
		{
			Name:       "ttl ignored day skip",
			Now:        timestamp(2006, time.January, 2, 15, 4), // Monday
			MinsToLive: 5,
			SkipDays:   []string{"Tuesday"},
			Want:       timestamp(2006, time.January, 2, 15, 9),
		},
		{
			Name:       "ttl skip one day",
			Now:        timestamp(2006, time.January, 2, 15, 4), // Monday
			MinsToLive: 5,
			SkipDays:   []string{"Monday"},
			Want:       timestamp(2006, time.January, 3, 0, 9),
		},
		{
			Name:       "ttl skip two days",
			Now:        timestamp(2006, time.January, 2, 15, 4), // Monday
			MinsToLive: 5,
			SkipDays:   []string{"Monday", "Tuesday", "Thursday"},
			Want:       timestamp(2006, time.January, 4, 0, 9),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			got := nextUpdate(test.Now, test.DefaultInterval, test.MinsToLive, test.SkipHours, test.SkipDays)
			if !got.Equal(test.Want) {
				const format = "2006 Jan 02  15:04:07 UTC"
				t.Fatalf("updateTime gave the wrong time:\nGot:  %s\nWant: %s", got.Format(format), test.Want.Format(format))
			}
		})
	}
}

const customLayout = "2006-01-02T15:04Z07:00"

var timeVal = time.Date(2015, 7, 1, 9, 27, 0, 0, time.UTC)

func TestParseTimeUsingOnlyDefaultLayouts(t *testing.T) {
	// Positive cases
	var singleLayout [1]string
	for _, layout := range defaultTimeLayouts {
		s := timeVal.Format(layout)
		singleLayout[0] = layout
		if tv := parseTime(s, singleLayout[:], nil); !tv.Equal(timeVal) {
			t.Errorf("expected times to equal, got %v", tv)
		}
	}

	// Negative cases
	if tv := parseTime("", defaultTimeLayouts, defaultTimeLayoutsWithNamedLocation); !tv.IsZero() {
		t.Error("expected error, got none")
	}

	if tv := parseTime("abc", defaultTimeLayouts, defaultTimeLayoutsWithNamedLocation); !tv.IsZero() {
		t.Error("expected error, got none")
	}

	custom := timeVal.Format(customLayout)
	if tv := parseTime(custom, defaultTimeLayouts, defaultTimeLayoutsWithNamedLocation); !tv.IsZero() {
		t.Error("expected error, got none")
	}
}

func TestParseTimeUsingCustomLayoutsPrepended(t *testing.T) {
	timeLayouts := append([]string{customLayout}, defaultTimeLayouts...)

	custom := timeVal.Format(customLayout)
	if tv := parseTime(custom, timeLayouts, defaultTimeLayoutsWithNamedLocation); !tv.Equal(timeVal) {
		t.Errorf("expected times to equal, got %v", tv)
	}
}

func TestParseWithTwoDigitYear(t *testing.T) {
	s := "Sun, 18 Dec 16 18:25:00 +0100"
	if tv := parseTime(s, defaultTimeLayouts, defaultTimeLayoutsWithNamedLocation); tv.Year() != 2016 {
		t.Errorf("expected year to be 2016, got year %v", tv.Year())
	}
}

// TestParseTime tests parseTime against some
// common and some not so valid time formts.
// Feel free to add more by adding another slice
// to the tests
func TestParseTime(t *testing.T) {
	tests := []struct {
		in       string
		expected time.Time
	}{
		{
			"Sun, 06 Sep 2009 16:20:00 +0000",
			time.Date(2009, 9, 6, 16, 20, 0, 0, time.UTC),
		},
		{
			"Sun, 06 Sep 2009 16:20:00",
			time.Date(2009, 9, 6, 16, 20, 0, 0, time.UTC),
		},
		{
			"Sun, 06 Sep 2009 16:20:00 -0300",
			time.Date(2009, 9, 6, 19, 20, 0, 0, time.UTC),
		},
		{
			"06 Sep 2009 16:18:00 EST",
			time.Date(2009, 9, 6, 21, 18, 0, 0, time.UTC),
		},
		{
			"06 Sep 2009 16:18:00",
			time.Date(2009, 9, 6, 16, 18, 0, 0, time.UTC),
		},
		{
			"Sun, 06 Sep 2009 16:18:00 EST",
			time.Date(2009, 9, 6, 21, 18, 0, 0, time.UTC),
		},
		{
			"Sun, 06 Sep 2010 16:43:59 EST +0100", // ignore EST
			time.Date(2010, 9, 6, 15, 43, 59, 0, time.UTC),
		},
		{
			"Sun, 06 Sep 2009 16:18:00 +0200 EST", // ignore EST
			time.Date(2009, 9, 6, 14, 18, 0, 0, time.UTC),
		},
		{
			"Sun, 06 Sep 2009 16:18:00 +0200",
			time.Date(2009, 9, 6, 14, 18, 0, 0, time.UTC),
		},
	}

	for _, c := range tests {
		tv := parseTime(c.in, defaultTimeLayouts, defaultTimeLayoutsWithNamedLocation)
		if tv.IsZero() {
			t.Errorf("%q Date is not valid", c.in)
		}

		if d := tv.UTC(); !d.Equal(c.expected) {
			t.Errorf("%q: %q != %q.", c.in, c.expected, d)
		}
	}
}
