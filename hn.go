package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	topStoriesURL = "https://hacker-news.firebaseio.com/v0/topstories.json"
	storyLinkURL  = "https://hacker-news.firebaseio.com/v0/item/%d.json"
	hnPostLinkURL = "https://news.ycombinator.com/item?id=%d"
)

func fetchTopStories(ctx context.Context, limit int) ([]*item, error) {
	// send items
	itemIds, err := fetchTopHNStories(ctx, topStoriesURL, limit)
	if err != nil {
		return nil, err
	}
	var items []*item
	for _, itID := range itemIds {
		if ctx.Err() != nil {
			return nil, nil
		}

		item, err := fetchItem(ctx, itID)
		if err != nil {
			log.Println(err) // warning
			continue
		}

		items = append(items, item)
	}

	return items, nil
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

func fetchTopHNStories(ctx context.Context, topStoriesURL string, limit int) ([]int, error) {
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
	var items []int
	err = decoder.Decode(&items)
	if err != nil {
		return nil, err
	}
	if len(items) < limit {
		limit = len(items)
	}

	return items[:limit], nil
}
