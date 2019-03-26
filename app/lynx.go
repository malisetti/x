package app

// Lynx contains the id of the url, output of the content visited by lynx and any error
type Lynx struct {
	ID     int
	Output string
	Err    error
}
