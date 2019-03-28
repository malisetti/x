package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/encrypt"
	"github.com/snabb/sitemap"
)

// JSONHandler serves /json
func JSONHandler(fetchItems func(since time.Time) ([]*app.Item, error), tstore *app.TempStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}
		t, err := strconv.Atoi(r.URL.Query().Get("t"))
		var since time.Time
		if err != nil || t <= 8 {
			since = time.Now().Add(-1 * app.EightHrs)
		} else if t > 8 && t <= 16 {
			since = time.Now().Add(-2 * app.EightHrs)
		} else {
			since = time.Now().Add(-3 * app.EightHrs)
		}

		items, err := fetchItems(since)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			log.Println(err)
		}
	}
}

// HTMLHandler serves /html
func HTMLHandler(fetchItems func(since time.Time) ([]*app.Item, error), tstore *app.TempStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t, err := strconv.Atoi(r.URL.Query().Get("t"))
		var since time.Time
		if err != nil || t <= 8 {
			since = time.Now().Add(-1 * app.EightHrs)
		} else if t > 8 && t <= 16 {
			since = time.Now().Add(-2 * app.EightHrs)
		} else {
			since = time.Now().Add(-3 * app.EightHrs)
		}

		items, err := fetchItems(since)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}

		tstore.RLock()
		defer tstore.RUnlock()
		data := make(map[string]interface{})
		data["bgColor"] = tstore.BgColor
		data["items"] = items
		err = tstore.Tmpl.Execute(w, data)
		if err != nil {
			log.Println(err)
		}
	}
}

// SitemapHandler serves sitemap.xml
func SitemapHandler(fetchItems func(since time.Time) ([]*app.Item, error), key *[32]byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eightHrsBack := time.Now().Add(-app.EightHrs)
		items, err := fetchItems(eightHrsBack)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}

		sm := sitemap.New()
		for _, it := range items {
			added := time.Unix(int64(it.Added), 0)
			h, err := encrypt.EncAndHex(it.URL, key)
			if err != nil {
				log.Println(err)
				continue
			}
			sm.Add(&sitemap.URL{
				Loc:        fmt.Sprintf("https://www.8hrs.xyz/l/%s", h),
				LastMod:    &added,
				ChangeFreq: sitemap.Hourly,
			})
		}

		_, err = sm.WriteTo(w)
		if err != nil {
			log.Println(err)
		}
	}
}

// FeedHandler serves rss|atom|json feeds from items
func FeedHandler(fetchItems func(since time.Time) ([]*app.Item, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		eightHrsBack := time.Now().Add(-app.EightHrs)
		items, err := fetchItems(eightHrsBack)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}

		vars := mux.Vars(r)
		feedType := vars["type"]
		now := time.Now()
		feed := &feeds.Feed{
			Title:       "Tech & News | Past 8hrs",
			Link:        &feeds.Link{Href: "https://8hrs.xyz"},
			Description: "Discussions about Technology, Science and other News from hackernews",
			Author:      &feeds.Author{Name: "Seshachalam Malisetti", Email: "abbiya@gmail.com"},
			Created:     now,
		}
		var feedItems []*feeds.Item
		for _, it := range items {
			feedItem := &feeds.Item{
				Title:   it.Title,
				Link:    &feeds.Link{Href: it.URL},
				Author:  &feeds.Author{Name: it.By},
				Created: time.Unix(int64(it.Added), 0),
			}
			if strings.TrimSpace(it.Description) != "" {
				feedItem.Description = it.Description
			} else if strings.TrimSpace(it.Textx) != "" {
				feedItem.Description = it.Textx
			} else {
				feedItem.Description = ""
			}
			feedItem.Id = strconv.Itoa(it.ID)
			feedItems = append(feedItems, feedItem)
		}

		feed.Items = feedItems

		switch feedType {
		case "atom":
			atom, err := feed.ToAtom()
			if err != nil {
				fmt.Fprintf(w, "%s", err)
			}

			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			fmt.Fprintf(w, "%s", atom)
			return
		case "rss":
			rss, err := feed.ToRss()
			if err != nil {
				fmt.Fprintf(w, "%s", err)
			}
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			fmt.Fprintf(w, "%s", rss)
			return
		default:
			// json
			j, err := feed.ToJSON()
			if err != nil {
				fmt.Fprintf(w, "%s", err)
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, "%s", j)
			return
		}
	}
}

// LinkHandler redirects encrypted links generated by 8hrs.xyz
func LinkHandler(key *[32]byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		h := vars["hash"]

		link, err := encrypt.DecFromHex(h, key)
		if err != nil {
			log.Println(err)
			if r.Method == http.MethodGet {
				http.NotFound(w, r)
				return
			}
		}

		if r.Method == http.MethodPost {
			log.Println(link)
			w.WriteHeader(http.StatusOK)
			return
		}

		log.Println(link)
		http.Redirect(w, r, link, http.StatusSeeOther)
	}
}

// FileHandler serves a file from a given path
func FileHandler(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}
