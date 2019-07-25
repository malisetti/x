package bringer

import (
	"context"
	"time"

	"github.com/mseshachalam/x/app"
)

// RandBringer brings random items
type RandBringer struct {
	Ctx context.Context
}

var items = []*app.Item{
	&app.Item{
		ID:          102,
		By:          "sesh",
		Title:       "sample rand",
		URL:         "8hrs.xyz",
		Deleted:     false,
		Dead:        false,
		DiscussLink: "rand",
		Added:       int(time.Now().Unix()),
	},
}

// SetContext sets context
func (b *RandBringer) SetContext(ctx context.Context) {
	b.Ctx = ctx
}

// Bring random news
func (b *RandBringer) Bring() ([]*app.Item, error) {
	return items, nil
}

// Fetch fetches items by ids
func (b *RandBringer) Fetch(ids []int) ([]*app.Item, error) {
	return items, nil
}

// GetURL makes an url for the given id
func (b *RandBringer) GetURL(id interface{}) string {
	return "rand"
}

// GetDiscussLink makes a discuss url for the given id
func (b *RandBringer) GetDiscussLink(id interface{}) string {
	return "rand"
}

// GetSource tells source identifier of items
func (b *RandBringer) GetSource() string {
	return "RAND"
}
