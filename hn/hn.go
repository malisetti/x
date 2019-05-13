// Package hn provides functions to fetch data from hacker news.
package hn

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/dbp"
	"github.com/mseshachalam/x/util"
)

const (
	// TopStoriesURL fetches top stories from HN
	TopStoriesURL = "https://hacker-news.firebaseio.com/v0/topstories.json"
	// StoryLinkURL fetches a story given its id
	StoryLinkURL = "https://hacker-news.firebaseio.com/v0/item/%d.json"
	// PostLinkURL is the HN URL for given id
	PostLinkURL = "https://news.ycombinator.com/item?id=%d"
)

// FetchHNStoriesOf fetches itsm from given ids
func FetchHNStoriesOf(ctx context.Context, ids []int) ([]*app.Item, error) {
	return FetchStoriesFrom(ctx, util.IntsToChan(ids))
}

// FetchStoriesFrom is a concurrent implementation of FetchHNStoriesOf
func FetchStoriesFrom(ctx context.Context, ids <-chan int) ([]*app.Item, error) {
	itemsCh := make(chan *app.Item)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range ids {
				if ctx.Err() != nil {
					return
				}
				item, err := FetchItem(ctx, id)
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

// FetchCurrentItems fetches items from db that are with given ids plus older items from since
func FetchCurrentItems(db *sql.DB, since time.Time, ids []int) ([]*app.Item, error) {
	oIds, err := dbp.SelectItemsIdsBefore(db, since.Unix())
	if err != nil {
		return nil, err
	}
	currentTopPlusEightHrs := append(ids, oIds...)

	var items []*app.Item
	ra := rand.New(rand.NewSource(time.Now().Unix()))
	num := ra.Intn(2)
	if num == 0 {
		items, err = dbp.SelectItemsByIDsDesc(db, currentTopPlusEightHrs)
	} else {
		items, err = dbp.SelectItemsByIDsAsc(db, currentTopPlusEightHrs)
	}

	return items, err
}

// FetchIds fetches item ids with top limit
func FetchIds(ctx context.Context, limit int) ([]int, error) {
	req, err := http.NewRequest(http.MethodGet, TopStoriesURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var ids []int
	err = decoder.Decode(&ids)
	if err != nil {
		return nil, err
	}
	if len(ids) < limit {
		limit = len(ids)
	}

	return ids[:limit], nil
}

// FetchItem fetches item with itemID from HN
func FetchItem(ctx context.Context, itemID int) (*app.Item, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(StoryLinkURL, itemID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var it *app.Item
	err = decoder.Decode(&it)
	if err != nil {
		return nil, err
	}

	return it, nil
}
