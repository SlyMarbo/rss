package rss

import (
	"io/ioutil"
	"log"
	"testing"
)

func Test_ParseTitle(t *testing.T) {
	m := map[string]string{
		//"test1":      "",
		//"test2":      "",
		"rss_0.92":   "Dave Winer: Grateful Dead",
		"rss_1.0":    "Golem.de",
		"rss_2.0":    "RSS Title",
		"rss_2.0-1":  "Liftoff News",
		"atom_1.0":   "Titel des Weblogs",
		"atom_1.0-1": "Golem.de",
	}

	for k, v := range m {
		d, e := ioutil.ReadFile("testdata/" + k)
		if e != nil {
			log.Print("Error when loading file ", k, ": ", e)
		}
		f, e := Parse(d)

		var o string
		if e == nil {
			o = f.Title
		}

		if o != v {
			log.Print("KEY: ", k)
			log.Print("ERROR: ", e)
			log.Print("GOT: '", o, "', EXPECTED: '", v, "'")
			t.Fail()
		}
	}
}
