package rss

var database *db
var disabled bool

func init() {
	database = new(db)
	database.req = make(chan string)
	database.res = make(chan bool)
	database.known = make(map[string]struct{})
	go database.Run()
}

type db struct {
	req   chan string
	res   chan bool
	known map[string]struct{}
}

func (d *db) Run() {
	for s := range d.req {
		if disabled {
			d.res <- false
		} else if _, ok := d.known[s]; ok {
			d.res <- true
		} else {
			d.known[s] = struct{}{}
			d.res <- false
		}
	}
}
