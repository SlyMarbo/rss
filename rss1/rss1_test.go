package rss1

import (
	"testing"

	"github.com/SlyMarbo/rss/v2/internal/testar"
)

func TestParse(t *testing.T) {
	tests := []string{
		"rss-0.91-1",
		"rss-1.0-1",
		"rss-1.0-enclosure-1",
		"rss-1.0-no-channel-1",
		"sample-rss-1-1",
		"sample-rss-1-modules-1",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			testar.Test(t, test, Parse, "github.com/SlyMarbo/rss/v2/rss1")
		})
	}
}
