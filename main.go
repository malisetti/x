package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"

	"net/http"
	"strings"
	"time"

	"sync"

	"github.com/gorilla/mux"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sim "github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/ChimeraCoder/anaconda"
	"github.com/gorilla/feeds"
	_ "github.com/mattn/go-sqlite3"
)

const (
	port     = "8080"
	eightHrs = 8 * 60 * 60 * time.Second
)

type tempStore struct {
	sync.RWMutex
	currentTop30ItemIds []int
	tmpl                *template.Template
}

var tstore tempStore

func main() {
	readTemplate := func() error {
		indexTemplatePath := os.Getenv("INDEX_TMPL_PATH")
		tmpl := template.New("index.html")
		tmpl, err := tmpl.ParseFiles(indexTemplatePath)
		tstore.Lock()
		defer tstore.Unlock()
		tstore.tmpl = tmpl
		return err
	}

	err := readTemplate()
	if err != nil {
		log.Println(err)
		return
	}

	twAccessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	twAccessTokenSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")
	twConsumerAPIKey := os.Getenv("TWITTER_CONSUMER_API_KEY")
	twConsumerSecretKey := os.Getenv("TWITTER_CONSUMER_SECRET_KEY")

	if twAccessToken == "" || twAccessTokenSecret == "" || twConsumerAPIKey == "" || twConsumerSecretKey == "" {
		log.Println("twitter tokens should be set in env")
		return
	}

	tapi := anaconda.NewTwitterApiWithCredentials(twAccessToken, twAccessTokenSecret, twConsumerAPIKey, twConsumerSecretKey)

	dbPath := os.Getenv("APP_DB_PATH")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Println(err)
		return
	}

	defer db.Close()

	store := sim.NewStore()
	// Define a limit rate to 5 requests per minute.
	rate, err := limiter.NewRateFromFormatted("5-M")
	if err != nil {
		log.Println(err)
		return
	}

	err = setupTables(db)
	if err != nil {
		log.Println(err)
		return
	}

	errs := updateItemsTable(db, addDescColumn, addImgsColumn, addTweetIDColumn)
	for _, err = range errs {
		log.Println(err)
	}

	rlMiddleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	flow(ctx, db, tapi)

	go func() {
		fiveMinTicker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				return
			case <-fiveMinTicker.C:
				flow(ctx, db, tapi)
			}
		}
	}()

	go func() {
		eightHrsTicker := time.NewTicker(eightHrs)
		for {
			select {
			case <-ctx.Done():
				return
			case <-eightHrsTicker.C:
				err := readTemplate()
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(os.Getenv("STATIC_DIR")))))

	r.Handle("/", rlMiddleware.Handler(withHeadersLogging(func(w http.ResponseWriter, r *http.Request) {
		var ids []int
		func() {
			tstore.RLock()
			defer tstore.RUnlock()
			ids = tstore.currentTop30ItemIds
		}()

		items, err := fetchCurrentItems(db, ids)

		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}

		requestedContentType := r.Header.Get("Content-Type")
		if requestedContentType == "application/json" {
			w.Header().Set("Content-Type", "application/json")
			err = json.NewEncoder(w).Encode(items)
			if err != nil {
				log.Println(err)
			}
		} else {
			func() {
				tstore.RLock()
				defer tstore.RUnlock()
				err = tstore.tmpl.Execute(w, items)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}))).Methods(http.MethodGet)

	r.Handle("/feed/{type}", rlMiddleware.Handler(withHeadersLogging(func(w http.ResponseWriter, r *http.Request) {
		var ids []int
		func() {
			tstore.RLock()
			defer tstore.RUnlock()
			ids = tstore.currentTop30ItemIds
		}()

		items, err := fetchCurrentItems(db, ids)

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
			if strings.TrimSpace(it.Text) != "" {
				feedItem.Description = it.Text
			} else if strings.TrimSpace(it.Descriprion) != "" {
				feedItem.Description = it.Descriprion
			} else {
				feedItem.Description = ""
			}
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
	})))

	http.Handle("/", r)

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = port
	}
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", httpPort),
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	log.Println(srv.ListenAndServe())
}

func withHeadersLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.Path)
		// log CF- headers
		for h, v := range r.Header {
			h = strings.ToUpper(h) // headers are case insensitive
			if strings.HasPrefix(h, "CF-") {
				log.Printf("%s : %s\n", strings.Replace(h, "CF-", "", -1), strings.Join(v, " "))
			}
		}
		log.Println(realIP(r))
		log.Println(r.UserAgent())

		next.ServeHTTP(w, r)
	}
}

func flow(ctx context.Context, db *sql.DB, tapi *anaconda.TwitterApi) {
	tctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	ids, err := fetchTopHNStories(ctx, 30)
	if err != nil {
		log.Println(err)
	} else {
		func() {
			tstore.Lock()
			defer tstore.Unlock()
			tstore.currentTop30ItemIds = ids
		}()
	}

	items, err := fetchTopStories(tctx, 30)
	if err != nil {
		log.Println(err)
	}

	sixteenHrsBack := time.Now().Add(-2 * eightHrs)

	resultItems, err := selectItemsBefore(db, sixteenHrsBack.Unix())
	if err != nil {
		log.Println(err)
	}
	var olderItemsIDsNotInTop []int
	var tweetIDsFromOlderItemsToBeDeleted []int64
	for _, it := range resultItems {
		there := false
		for _, topIt := range items {
			if it.ID == topIt.ID {
				there = true
				break
			}
		}
		if !there {
			olderItemsIDsNotInTop = append(olderItemsIDsNotInTop, it.ID)
			if it.TweetID != 0 {
				tweetIDsFromOlderItemsToBeDeleted = append(tweetIDsFromOlderItemsToBeDeleted, it.TweetID)
			}
		}
	}
	if len(olderItemsIDsNotInTop) > 0 {
		err = deleteItemsWith(db, olderItemsIDsNotInTop)
		if err != nil {
			log.Println(err)
		}

		errs := deleteTweets(ctx, tapi, tweetIDsFromOlderItemsToBeDeleted)
		for id, err := range errs {
			log.Printf("%d tweet deletion failed with %s\n", id, err)
		}
	}

	eightHrsBack := time.Now().Add(-1 * eightHrs)
	resultItems, err = selectItemsAfter(db, eightHrsBack.Unix())
	if err != nil {
		log.Println(err)
	}
	var olderItemsIDsInTop []int
	for _, it := range resultItems {
		there := false
		for _, topIt := range items {
			if it.ID == topIt.ID {
				there = true
				break
			}
		}
		if there {
			olderItemsIDsInTop = append(olderItemsIDsInTop, it.ID)
		}
	}

	updatedItems, err := fetchHNStoriesOf(ctx, olderItemsIDsInTop)
	if err != nil {
		log.Println()
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

	// err = populateItemsWithPreview(items)
	// if err != nil {
	// 	log.Println(err)
	// }

	var itemIDs []int
	for _, it := range items {
		itemIDs = append(itemIDs, it.ID)
	}

	idToTweetIDs, err := fetchTweetIDsFor(db, itemIDs)
	if err != nil {
		log.Println(err)
	}
	for _, it := range items {
		if tweetID, ok := idToTweetIDs[it.ID]; ok {
			it.TweetID = tweetID
		}

		if it.URL == "" {
			it.URL = fmt.Sprintf(hnPostLinkURL, it.ID)
		}
		it.DiscussLink = fmt.Sprintf(hnPostLinkURL, it.ID)
		domain, err := urlToDomain(it.URL)
		if err == nil {
			it.Domain = domain
		}
	}

	errs := tweetItems(ctx, tapi, items)
	for id, err := range errs {
		log.Printf("%d tweeting failed with %s\n", id, err)
	}
	_, err = insertOrReplaceItems(db, items)
	if err != nil {
		log.Println(err)
	}
}
