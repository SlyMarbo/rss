package rss

var database *db
var disabled bool

func init() {
	database = NewDB()
	go database.Run()
}

type db struct {
	req   chan string
	res   chan bool
	known map[string]struct{}
}

func (d *db) Run() {
	d.known = make(map[string]struct{})
	var s string

	for {
		s = <-d.req
		if disabled {
			d.res <- false
		} else if _, ok := d.known[s]; ok {
			d.res <- true
		} else {
			d.res <- false
			d.known[s] = struct{}{}
		}
	}
}

func NewDB() *db {
	out := new(db)
	out.req = make(chan string)
	out.res = make(chan bool)
	return out
}
