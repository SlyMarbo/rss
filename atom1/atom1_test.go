package atom1

import (
	"testing"

	"github.com/SlyMarbo/rss/v2/internal/testar"
)

func TestParse(t *testing.T) {
	tests := []string{
		"atom-1.0-1",
		"atom-1.0-2",
		"atom-1.0-enclosure-1",
		"atom-1.0-html-1",
		"feedparser",
		"google-merchant-center",
		"rss-2.0-1",
	}

	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			testar.Test(t, test, Parse, "github.com/SlyMarbo/rss/v2/atom1")
		})
	}
}
