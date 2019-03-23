package app

import (
	"html/template"
	"sync"
)

// TempStore is temp storage for app
type TempStore struct {
	sync.RWMutex
	// CurrentTop30ItemIds is the list of top items.
	CurrentTop30ItemIds []int
	// Tmpl is the template.
	Tmpl *template.Template
	// BgColor of the rendered template.
	BgColor string
}
