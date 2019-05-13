package app

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
