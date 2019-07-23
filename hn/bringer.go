package hn

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/util"
)

// HackerNewsBringer brings news from hn
type HackerNewsBringer struct {
	NumberOfItems int
	Ctx           context.Context
}

// Bring hn news
func (hnb *HackerNewsBringer) Bring() ([]*app.Item, error) {
	itemsCh := make(chan *app.Item)
	ids, err := FetchIds(hnb.Ctx, hnb.NumberOfItems)
	if err != nil {
		return nil, err
	}
	idsCh := util.IntsToChan(ids)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idsCh {
				if hnb.Ctx.Err() != nil {
					return
				}
				item, err := FetchItem(hnb.Ctx, id)
				if err != nil {
					log.Println(err) // warning
					continue
				}

				itemsCh <- item
			}
		}()
	}

	go func() {
		wg.Wait()
		close(itemsCh)
	}()

	var items []*app.Item
	for it := range itemsCh {
		items = append(items, it)
	}

	return items, nil
}

// HackerNewsPeriodicBringer implements Periodic Bringer
type HackerNewsPeriodicBringer struct {
	Ctx      context.Context
	Interval time.Duration
}

// Bring gives a hn bringer periodically
func (hnpb *HackerNewsPeriodicBringer) Bring() <-chan app.Bringer {
	out := make(chan app.Bringer)
	go func() {
		hnb := new(HackerNewsBringer)
		hnb.NumberOfItems = 30
		hnb.Ctx = hnpb.Ctx
		out <- hnb

		ticker := time.NewTicker(hnpb.Interval)
		for {
			select {
			case <-ticker.C:
				hnb := new(HackerNewsBringer)
				hnb.NumberOfItems = 30
				hnb.Ctx = hnpb.Ctx
				out <- hnb
			case <-hnpb.Ctx.Done():
				close(out)
				return
			}
		}
	}()

	return out
}
