package app

import (
	"html/template"
	"sync"
)

// TempStore is temp storage for app
type TempStore struct {
	sync.RWMutex
	CurrentTop30ItemIds []int
	Tmpl                *template.Template
	BgColor             string
}
