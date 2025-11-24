package rss2

import (
	"testing"

	"github.com/SlyMarbo/rss/v2/internal/testar"
)

func TestParse(t *testing.T) {
	tests := []string{
		"rss-0.91-1",
		"rss-0.92-1",
		"rss-1.0-1",
		"rss-2.0-1",
		"rss-2.0-2",
		"rss-2.0-content-encoded-1",
		"rss-2.0-enclosure-1",
		"rss-2.0-enclosure-2",
		"rss-2.0-no-channel-1",
		"rss-2.0-update-1",
		"rss-2.0-update-2",
		"sample-rss-2-1",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			testar.Test(t, test, Parse, "github.com/SlyMarbo/rss/v2/rss2")
		})
	}
}
