package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	topStoriesURL = "https://hacker-news.firebaseio.com/v0/topstories.json"
	storyLinkURL  = "https://hacker-news.firebaseio.com/v0/item/%d.json"
	hnPostLinkURL = "https://news.ycombinator.com/item?id=%d"
)

var r = strings.NewReplacer("http://", "", "https://", "", "www.", "", "www2.", "", "www3.", "")

func fetchCurrentItems(db *sql.DB, ids []int) ([]*item, error) {
	eightHrsBack := time.Now().Add(-eightHrs)
	oIds, err := selectItemsIdsBefore(db, eightHrsBack.Unix())
	if err != nil {
		return nil, err
	}
	currentTopPlusEightHrs := append(ids, oIds...)

	var items []*item
	ra := rand.New(rand.NewSource(time.Now().Unix()))
	num := ra.Intn(2)
	if num == 0 {
		items, err = selectItemsByIDsDesc(db, currentTopPlusEightHrs)
	} else {
		items, err = selectItemsByIDsAsc(db, currentTopPlusEightHrs)
	}

	return items, err
}

func urlToDomain(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Hostname(), ".")
	if len(parts) >= 2 {
		domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
		return domain, nil
	}

	return r.Replace(u.Hostname()), nil
}

func fetchHNStoriesOf(ctx context.Context, ids []int) ([]*item, error) {
	return fetchStoriesFrom(ctx, intsToChan(ids))
}

func fetchStoriesFrom(ctx context.Context, ids <-chan int) ([]*item, error) {
	itemsCh := make(chan *item)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range ids {
				item, err := fetchItem(ctx, id)
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

	var items []*item
	for it := range itemsCh {
		items = append(items, it)
	}

	return items, nil
}

func intsToChan(itemIds []int) <-chan int {
	ids := make(chan int)
	go func() {
		defer close(ids)
		for _, itID := range itemIds {
			ids <- itID
		}
	}()
	return ids
}

func fetchTopStories(ctx context.Context, limit int) ([]*item, error) {
	// send items
	itemIds, err := fetchTopHNStories(ctx, limit)
	if err != nil {
		return nil, err
	}

	return fetchStoriesFrom(ctx, intsToChan(itemIds))
}

func fetchItem(ctx context.Context, itemID int) (*item, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf(storyLinkURL, itemID), nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var it *item
	err = decoder.Decode(&it)
	if err != nil {
		return nil, err
	}

	return it, nil
}

func fetchTopHNStories(ctx context.Context, limit int) ([]int, error) {
	req, err := http.NewRequest(http.MethodGet, topStoriesURL, nil)
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
