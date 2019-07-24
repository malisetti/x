package hn

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/dbp"
	"github.com/mseshachalam/x/encrypt"
	"github.com/mseshachalam/x/util"
)

// Maintainer implements Maintainer
type Maintainer struct {
	Ctx             context.Context
	Config          *app.Config
	PediodicBringer *PeriodicBringer
	Storage         *sql.DB
	Key             *[32]byte
}

// Maintain takes care of storage and updates to items
func (m *Maintainer) Maintain() {
	for bringer := range m.PediodicBringer.Bring() {
		items, err := bringer.Bring()
		if err != nil {
			log.Println(err)
		}

		var ids []int
		for _, item := range items {
			ids = append(ids, item.ID)
		}
		// Update items to latest timestamp
		err = dbp.UpdateItemsAddedTimeToNow(m.Storage, ids)
		if err != nil {
			log.Println(err)
		}

		thirtyTwoHrsBack := time.Now().Add(-4 * app.EightHrs)
		olderItemsIDsNotInTop, err := dbp.SelectItemsIdsBeforeAndNotOf(m.Storage, thirtyTwoHrsBack.Unix(), ids)
		if err != nil {
			log.Println(err)
		}

		err = dbp.DeleteItemsWith(m.Storage, olderItemsIDsNotInTop)
		if err != nil {
			log.Println(err)
		}

		eightHrsBack := time.Now().Add(-1 * app.EightHrs)
		olderItemsIDsInTop, err := dbp.SelectItemsIDsAfterAndNotOf(m.Storage, eightHrsBack.Unix(), ids)
		if err != nil {
			log.Println(err)
		}

		updatedItems, err := FetchHNStoriesOf(m.Ctx, olderItemsIDsInTop)
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
				ids = append(ids, updatedItem.ID)
			}
		}

		existingItems, err := dbp.SelectExistingPropsOfItemsByIDsAsc(m.Storage, ids)
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
				it.Description = eit.Description
				it.EncryptedURL = eit.EncryptedURL
				it.EncryptedDiscussLink = eit.EncryptedDiscussLink

				break
			}
		}

		idsToURLs := make(map[int]string)
		// visit the link with lynx and update description
		for _, it := range items {
			if it.Description == "" {
				idsToURLs[it.ID] = it.URL
			}

			if it.URL == "" {
				it.URL = fmt.Sprintf(PostLinkURL, it.ID)
			}
			if it.DiscussLink == "" {
				it.DiscussLink = fmt.Sprintf(PostLinkURL, it.ID)
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
				h, _ := encrypt.EncAndHex(link, m.Key)
				it.EncryptedURL = h
			}
			if it.EncryptedDiscussLink == "" {
				h, _ := encrypt.EncAndHex(it.DiscussLink, m.Key)
				it.EncryptedDiscussLink = h
			}
		}

		err = dbp.InsertOrReplaceItems(m.Storage, items)
		if err != nil {
			log.Println(err)
		}
	}
}
