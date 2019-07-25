package app

import "context"

// PeriodicBringer gives us a bringer periodically
type PeriodicBringer interface {
	Bring() <-chan Bringer
}

// Bringer brings a list of things from other portals
type Bringer interface {
	SetContext(context.Context)
	Bring() ([]*Item, error)
	Fetch([]int) ([]*Item, error)
	GetURL(interface{}) string
	GetDiscussLink(interface{}) string
	GetSource() string
}
