// go run debug.go [URL]

// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SlyMarbo/rss"
)

func main() {
	if len(os.Args) != 2 || os.Args[1] == "-h" {
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [URL]\n", filepath.Base(os.Args[0]))
		os.Exit(2)
	}

	feed, err := rss.Fetch(os.Args[1])
	if err != nil {
		panic(err)
	}

	raw, err := json.Marshal(feed)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	if err := json.Indent(buf, raw, "", "\t"); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}
