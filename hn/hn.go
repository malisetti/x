package hn

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/exec"
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

// FetchHNStoriesOf fetches itsm from given ids
func FetchHNStoriesOf(ctx context.Context, ids []int) ([]*app.Item, error) {
	return fetchStoriesFrom(ctx, util.IntsToChan(ids))
}

func fetchStoriesFrom(ctx context.Context, ids <-chan int) ([]*app.Item, error) {
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

// FetchTopStories fetches items from ids if not empty or fetches limit stories from HN
func FetchTopStories(ctx context.Context, ids []int, limit int) ([]*app.Item, error) {
	// send items
	var itemIds []int
	if len(ids) == 0 {
		var err error
		itemIds, err = FetchTopHNStories(ctx, limit)
		if err != nil {
			return nil, err
		}
	} else {
		itemIds = ids
	}

	return fetchStoriesFrom(ctx, util.IntsToChan(itemIds))
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

// FetchTopHNStories fetches stories ids from HN with given limit
func FetchTopHNStories(ctx context.Context, limit int) ([]int, error) {
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

// TODO: Do not use
func visitAndGetDescription(ctx context.Context, idsToURLs map[int]string) map[int]string {
	out := make(map[int]string)
	var wg sync.WaitGroup
	limit := 0
	for i, u := range idsToURLs {
		if limit < 4 {
			wg.Add(1)
			go func(i int, u string) (b []byte, err error) {
				defer func() {
					wg.Done()
					if err != nil {
						out[i] = string(b)
					}
					limit--
				}()

				cmd := exec.CommandContext(ctx, "lynx", "-dump", u)
				b, err = cmd.Output()
				return
			}(i, u)
		}
	}
	wg.Wait()

	return out
}
