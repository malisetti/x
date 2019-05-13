package app

import (
	"time"
)

// PeriodicBringer gives us a bringer periodically
type PeriodicBringer interface {
	Bring() <-chan Bringer
}

// Bringer brings a list of things from other portals
type Bringer interface {
	Bring() ([]*Item, error)
}

// Maintainer maintains items and storage
type Maintainer interface {
	Maintain()
}

// Options are query options
type Options struct {
	Period time.Duration
}

// Option is something
type Option func(*Options)

// Querier queries items which the bringer got
type Querier interface {
	Query(opts ...Option) []Item
}
