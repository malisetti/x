package twitter

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/util"
)

const tweetStatus = "%s    (%s)\n\n%s"

var hnReplacer = strings.NewReplacer("Show HN:", "", "Ask HN:", "")

// TweetItems tweets given items using anaconda twitter client
func TweetItems(ctx context.Context, tapi *anaconda.TwitterApi, items []*app.Item) map[int64]error {
	errs := make(map[int64]error)
	itemsChan := util.ItemsToChan(items)
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

				link := fmt.Sprintf("https://www.8hrs.xyz/l/%s", it.EncryptedURL)
				status := fmt.Sprintf(tweetStatus, it.Title, strings.Replace(it.Domain, ".", " dot ", -1), link)
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

// DeleteTweets deletes tweets with given ids using anaconda twitter client
func DeleteTweets(ctx context.Context, tapi *anaconda.TwitterApi, ids []int64) map[int64]error {
	errs := make(map[int64]error)
	idCh := util.Int64sToChan(ids)
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
