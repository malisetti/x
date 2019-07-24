package hn

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/util"
)

// Bringer brings news from hn
type Bringer struct {
	NumberOfItems int
	Ctx           context.Context
	NWorkers      int
}

// Bring hn news
func (b *Bringer) Bring() ([]*app.Item, error) {
	itemsCh := make(chan *app.Item)
	ids, err := FetchIds(b.Ctx, b.NumberOfItems)
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
				item, err := FetchItem(b.Ctx, id)
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
		b := new(Bringer)
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
