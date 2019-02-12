package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ChimeraCoder/anaconda"
)

const tweetStatus = "%s\n%s"

var hnReplacer = strings.NewReplacer("Show HN:", "", "Ask HN:", "")

func tweetItems(ctx context.Context, tapi *anaconda.TwitterApi, items []*item) map[int64]error {
	errs := make(map[int64]error)
	itemsChan := itemsToChan(items)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for it := range itemsChan {
				if ctx.Err() != nil {
					return
				}
				if it.TweetID > 0 {
					continue
				}

				link := it.URL
				if link == "" {
					link = it.DiscussLink
				}
				status := fmt.Sprintf(tweetStatus, it.Title, link)
				status = hnReplacer.Replace(status)
				status = strings.TrimSpace(status)
				tweet, err := tapi.PostTweet(status, nil)
				if err != nil {
					errs[it.TweetID] = err
					continue
				}

				it.TweetID = tweet.Id
			}
		}()
	}
	wg.Wait()

	return errs
}

func deleteTweets(ctx context.Context, tapi *anaconda.TwitterApi, ids []int64) map[int64]error {
	errs := make(map[int64]error)
	idCh := int64sToChan(ids)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tweetID := range idCh {
				if ctx.Err() != nil {
					return
				}

				_, err := tapi.DeleteTweet(tweetID, false)
				if err != nil {
					errs[tweetID] = err
				}
			}
		}()
	}
	wg.Wait()

	return errs
}
