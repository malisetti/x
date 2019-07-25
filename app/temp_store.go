package app

import (
	"html/template"
	"sync"
)

// TempStore is temp storage for app
type TempStore struct {
	sync.RWMutex
	// Tmpl is the template.
	Tmpl *template.Template
	// BgColor of the rendered template.
	BgColor string
}
