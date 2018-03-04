package rss

import (
	"testing"
	"time"
)

const customLayout = "2006-01-02T15:04Z07:00"

var (
	timeVal         = time.Date(2015, 7, 1, 9, 27, 0, 0, time.UTC)
	originalLayouts = TimeLayouts
)

func TestParseTimeUsingOnlyDefaultLayouts(t *testing.T) {
	// Positive cases
	for _, layout := range originalLayouts {
		s := timeVal.Format(layout)
		if tv, err := parseTime(s); err != nil || !tv.Equal(timeVal) {
			t.Errorf("expected no err and times to equal, got err %v and time value %v", err, tv)
		}
	}

	// Negative cases
	if _, err := parseTime(""); err == nil {
		t.Error("expected err, got none")
	}

	if _, err := parseTime("abc"); err == nil {
		t.Error("expected err, got none")
	}

	custom := timeVal.Format(customLayout)
	if _, err := parseTime(custom); err == nil {
		t.Error("expected err, got none")
	}
}

func TestParseTimeUsingCustomLayoutsPrepended(t *testing.T) {
	TimeLayouts = append([]string{customLayout}, originalLayouts...)

	custom := timeVal.Format(customLayout)
	if tv, err := parseTime(custom); err != nil || !tv.Equal(timeVal) {
		t.Errorf("expected no err and times to equal, got err %v and time value %v", err, tv)
	}

	TimeLayouts = originalLayouts
}

func TestParseTimeUsingCustomLayoutsAppended(t *testing.T) {
	TimeLayouts = append(originalLayouts, customLayout)

	custom := timeVal.Format(customLayout)
	if tv, err := parseTime(custom); err != nil || !tv.Equal(timeVal) {
		t.Errorf("expected no err and times to equal, got err %v and time value %v", err, tv)
	}

	TimeLayouts = originalLayouts
}

func TestParseWithTwoDigitYear(t *testing.T) {
	s := "Sun, 18 Dec 16 18:25:00 +0100"
	if tv, err := parseTime(s); err != nil || tv.Year() != 2016 {
		t.Errorf("expected no err and year to be 2016, got err %v, and year %v", err, tv.Year())
	}
}
