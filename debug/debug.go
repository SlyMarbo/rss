// Run with: go run -tags debug debug.go [URL]
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SlyMarbo/rss"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [URL]\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	feed, err := rss.Fetch(os.Args[1])
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", feed)
}
