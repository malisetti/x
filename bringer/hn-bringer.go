package bringer

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/hn"
	"github.com/mseshachalam/x/util"
)

// HNBringer brings news from hn
type HNBringer struct {
	NumberOfItems int
	Ctx           context.Context
	NWorkers      int
}

// Fetch fetches items by ids
func (b *HNBringer) Fetch(ids []int) ([]*app.Item, error) {
	return hn.FetchHNStoriesOf(b.Ctx, ids)
}

// GetURL makes an url for the given id
func (b *HNBringer) GetURL(id interface{}) string {
	return fmt.Sprintf(hn.PostLinkURL, id)
}

// GetDiscussLink makes a discuss url for the given id
func (b *HNBringer) GetDiscussLink(id interface{}) string {
	return fmt.Sprintf(hn.PostLinkURL, id)
}

// GetSource tells source identifier of items
func (b *HNBringer) GetSource() string {
	return "HN"
}

// Bring hn news
func (b *HNBringer) Bring() ([]*app.Item, error) {
	itemsCh := make(chan *app.Item)
	ids, err := hn.FetchIds(b.Ctx, b.NumberOfItems)
	if err != nil {
		return nil, err
	}
	idsCh := util.IntsToChan(ids)
	var wg sync.WaitGroup
	for i := 0; i < b.NWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idsCh {
				if b.Ctx.Err() != nil {
					break
				}
				item, err := hn.FetchItem(b.Ctx, id)
				if err != nil {
					log.Println(err) // warning
					continue
				}

				itemsCh <- item
			}
		}()
	}

	go func() {
		defer close(itemsCh)
		wg.Wait()
	}()

	var items []*app.Item
	for it := range itemsCh {
		items = append(items, it)
	}

	return items, nil
}

// PeriodicBringer implements Periodic Bringer
type PeriodicBringer struct {
	Ctx      context.Context
	Interval time.Duration
}

// Bring gives a hn bringer periodically
func (pb *PeriodicBringer) Bring() <-chan app.Bringer {
	out := make(chan app.Bringer)
	go func() {
		defer close(out)
		b := new(HNBringer)
		b.NumberOfItems = app.DefaultHNFrontPageArticlesCount
		b.Ctx = pb.Ctx
		b.NWorkers = 4

		out <- b

		ticker := time.NewTicker(pb.Interval)
		for {
			select {
			case <-ticker.C:
				out <- b
			case <-pb.Ctx.Done():
				break
			}
		}
	}()

	return out
}
