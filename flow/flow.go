// Package flow is the logic of the app.
package flow

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ChimeraCoder/anaconda"
	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/dbp"
	"github.com/mseshachalam/x/encrypt"
	"github.com/mseshachalam/x/hn"
	"github.com/mseshachalam/x/twitter"
	"github.com/mseshachalam/x/util"
)

// Flow is the logic of application
func Flow(ctx context.Context, tstore *app.TempStore, db *sql.DB, conf *app.Config, tapi *anaconda.TwitterApi, key *[32]byte) {
	ids, err := hn.FetchTopHNStories(ctx, 30)
	if err != nil {
		log.Println(err)
	} else {
		func() {
			tstore.Lock()
			defer tstore.Unlock()
			tstore.CurrentTop30ItemIds = ids
		}()
	}

	items, err := hn.FetchTopStories(ctx, ids, 30)
	if err != nil {
		log.Println(err)
	}

	thirtyTwoHrsBack := time.Now().Add(-4 * app.EightHrs)

	err = dbp.UpdateItemsAddedTimeToNow(db, ids)
	if err != nil {
		log.Println(err)
	}

	idAndTweetIDs, err := dbp.SelectItemsIDsBefore(db, thirtyTwoHrsBack.Unix())
	if err != nil {
		log.Println(err)
	}
	var olderItemsIDsNotInTop []int
	var tweetIDsFromOlderItemsToBeDeleted []int64
	for id, tid := range idAndTweetIDs {
		there := false
		for _, topIt := range items {
			if id == topIt.ID {
				there = true
				break
			}
		}
		if !there {
			olderItemsIDsNotInTop = append(olderItemsIDsNotInTop, id)
			if tid != 0 {
				tweetIDsFromOlderItemsToBeDeleted = append(tweetIDsFromOlderItemsToBeDeleted, tid)
			}
		}
	}
	if len(olderItemsIDsNotInTop) > 0 {
		err = dbp.DeleteItemsWith(db, olderItemsIDsNotInTop)
		if err != nil {
			log.Println(err)
		}

		if conf.TweetItems {
			errs := twitter.DeleteTweets(ctx, tapi, tweetIDsFromOlderItemsToBeDeleted)
			for id, err := range errs {
				log.Printf("%d tweet deletion failed with %s\n", id, err)
			}
		}
	}

	eightHrsBack := time.Now().Add(-1 * app.EightHrs)
	itemsIDsFromLastEightHrs, err := dbp.SelectItemsIDsAfter(db, eightHrsBack.Unix())
	if err != nil {
		log.Println(err)
	}
	var olderItemsIDsInTop []int
	for _, id := range itemsIDsFromLastEightHrs {
		there := false
		for _, topIt := range items {
			if id == topIt.ID {
				there = true
				break
			}
		}
		if there {
			olderItemsIDsInTop = append(olderItemsIDsInTop, id)
		}
	}

	updatedItems, err := hn.FetchHNStoriesOf(ctx, olderItemsIDsInTop)
	if err != nil {
		log.Println(err)
	}

	for _, updatedItem := range updatedItems {
		there := false
		for _, it := range items {
			if it.ID == updatedItem.ID {
				there = true
				break
			}
		}
		if !there {
			items = append(items, updatedItem)
		}
	}

	var itemIDs []int
	for _, it := range items {
		itemIDs = append(itemIDs, it.ID)
	}

	existingItems, err := dbp.SelectExistingPropsOfItemsByIDsAsc(db, itemIDs)
	if err != nil {
		log.Println(err)
	}
	for _, eit := range existingItems {
		for _, it := range items {
			if eit.ID != it.ID {
				continue
			}

			it.URL = eit.URL
			it.DiscussLink = eit.DiscussLink
			it.Domain = eit.Domain
			it.TweetID = eit.TweetID
			it.Description = eit.Description
			it.EncryptedURL = eit.EncryptedURL
			it.EncryptedDiscussLink = eit.EncryptedDiscussLink

			break
		}
	}

	// visit the link with lynx and update description
	for _, it := range items {
		if it.Description == "" {

		}

		if it.URL == "" {
			it.URL = fmt.Sprintf(hn.PostLinkURL, it.ID)
		}
		if it.DiscussLink == "" {
			it.DiscussLink = fmt.Sprintf(hn.PostLinkURL, it.ID)
		}
		if it.Domain == "" {
			domain, err := util.URLToDomain(it.URL)
			if err == nil {
				it.Domain = domain
			}
		}
		if it.EncryptedURL == "" {
			link := it.URL
			if link == "" {
				link = it.DiscussLink
			}
			h, _ := encrypt.EncAndHex(link, key)
			it.EncryptedURL = h
		}
		if it.EncryptedDiscussLink == "" {
			h, _ := encrypt.EncAndHex(it.DiscussLink, key)
			it.EncryptedDiscussLink = h
		}
	}

	if conf.TweetItems {
		errs := twitter.TweetItems(ctx, tapi, items)
		for id, err := range errs {
			log.Printf("%d tweeting failed with %s\n", id, err)
		}
	}

	err = dbp.InsertOrReplaceItems(db, items)
	if err != nil {
		log.Println(err)
	}

	if conf.PingGoogle {
		_, err = http.Get(fmt.Sprintf("https://www.google.com/ping?sitemap=%s", "https://www.8hrs.xyz/sitemap.xml"))
		if err != nil {
			log.Println(err)
		}
	}
}
