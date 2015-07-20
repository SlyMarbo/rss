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
			t.Error("expected no err and times to equal, got err %v and time value %v", err, tv)
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
		t.Error("expected no err and times to equal, got err %v and time value %v", err, tv)
	}
	TimeLayouts = originalLayouts
}

func TestParseTimeUsingCustomLayoutsAppended(t *testing.T) {
	TimeLayouts = append(originalLayouts, customLayout)
	custom := timeVal.Format(customLayout)
	if tv, err := parseTime(custom); err != nil || !tv.Equal(timeVal) {
		t.Error("expected no err and times to equal, got err %v and time value %v", err, tv)
	}
	TimeLayouts = originalLayouts
}
