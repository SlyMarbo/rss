package rss

import "testing"

func TestShouldStripImageFromContent(t *testing.T) {
	var data = []struct {
		orig string
		exp  string
	}{
		{"Some Content<img src=\"http://link/to/image.jpg\"/>", "Some Content"},
		{"Before <img src=\"http://link/to/image.jpg\"/> After", "Before  After"},
		{"<img src=\"http://link/to/image.jpg\"/>Image was at the beginning.", "Image was at the beginning."},
		{"Image had no <img src=\"http://link/to/image.jpg\"/>slash at the end of the tag", "Image had no slash at the end of the tag"},
		{"All <img src=\"http://link/to/image.jpg\"/>images <img src=\"http://link/to/image.jpg\"/>are <img src=\"http://link/to/image.jpg\"/>gone", "All images are gone"},
	}

	for _, d := range data {
		it := Item{Title: "Title", Content: d.orig}

		c := it.RawContent()

		if c != d.exp {
			t.Errorf("Raw Content expected: %v, got %v", d.exp, c)
		}
	}

}
