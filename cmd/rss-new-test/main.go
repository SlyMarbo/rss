// Command rss-new-test is a helper for recording new test vectors.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SlyMarbo/rss/v2"
	"github.com/SlyMarbo/rss/v2/atom1"
	"github.com/SlyMarbo/rss/v2/rss1"
	"github.com/SlyMarbo/rss/v2/rss2"

	"golang.org/x/tools/txtar"
)

func init() {
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
	log.SetPrefix("")
}

func main() {
	var help, doAtom1, doRSS1, doRSS2 bool
	var outName string

	flag.BoolVar(&help, "h", false, "Print this message and exit.")
	flag.BoolVar(&doAtom1, "atom1", false, "Output in atom1 format.")
	flag.BoolVar(&doRSS1, "rss1", false, "Output in rss1 format.")
	flag.BoolVar(&doRSS2, "rss2", false, "Output in rss2 format.")
	flag.StringVar(&outName, "o", "", "The filename for the new test vector.")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s OPTIONS... URL\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.Parse()
	if help || outName == "" || flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if trues, _ := countBools(doAtom1, doRSS1, doRSS2); trues > 1 {
		fmt.Fprintf(os.Stderr, "Can use at most one of -atom1, -rss1, and -rss2.\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Fetch the URL.
	url := flag.Arg(0)
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("Failed to fetch feed: %v", err)
	}

	payload, err := io.ReadAll(res.Body)
	if err != nil {
		res.Body.Close()
		log.Fatalf("Failed to read feed content: %v", err)
	}

	if err = res.Body.Close(); err != nil {
		// Warn but continue.
		log.Printf("Failed to close feed content: %v", err)
	}

	var feed any
	var format string
	switch {
	case doAtom1:
		format = "github.com/SlyMarbo/rss/v2/atom1"
		feed, err = atom1.Parse(bytes.NewReader(payload))
	case doRSS1:
		format = "github.com/SlyMarbo/rss/v2/rss1"
		feed, err = rss1.Parse(bytes.NewReader(payload))
	case doRSS2:
		format = "github.com/SlyMarbo/rss/v2/rss2"
		feed, err = rss2.Parse(bytes.NewReader(payload))
	default:
		r, err := rss.NewReader()
		if err != nil {
			log.Fatalf("Failed to construct RSS reader: %v", err)
		}

		format = "github.com/SlyMarbo/rss/v2"
		feed, err = r.Parse(url, bytes.NewReader(payload))
	}
	if err != nil {
		res.Body.Close()
		log.Fatalf("Failed to parse feed: %v", err)
	}

	enc, err := json.MarshalIndent(feed, "", "\t")
	if err != nil {
		log.Fatalf("Failed to encode feed: %v", err)
	}

	archive := &txtar.Archive{
		Files: []txtar.File{
			{
				Name: "URL",
				Data: []byte(url),
			},
			{
				Name: "format",
				Data: []byte(format),
			},
			{
				Name: "XML",
				Data: payload,
			},
			{
				Name: "JSON",
				Data: enc,
			},
		},
	}

	vector := txtar.Format(archive)
	err = os.WriteFile(outName, vector, 0666)
	if err != nil {
		log.Fatalf("Failed to write vector: %v", err)
	}
}

func countBools(v ...bool) (trues, falses int) {
	for _, v := range v {
		if v {
			trues++
		} else {
			falses++
		}
	}

	return trues, falses
}
