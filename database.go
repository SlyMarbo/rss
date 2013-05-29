package rss

type Seen map[string]struct{}

func NewSeen() Seen {
	return map[string]struct{}{}
}
