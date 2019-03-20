package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"os"
	"strconv"

	"net/http"
	"strings"
	"time"

	"sync"

	"github.com/gorilla/mux"
	"github.com/snabb/sitemap"

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

var bgColors = []string{"aliceblue", "antiquewhite", "aqua", "aquamarine", "azure", "beige", "bisque", "blanchedalmond", "burlywood", "chartreuse"}

type tempStore struct {
	sync.RWMutex
	currentTop30ItemIds []int
	tmpl                *template.Template
	bgColor             string
}

var tstore tempStore

func main() {
	var conf *config
	configFilePath := os.Getenv("APP_CONFIG_PATH")
	f, err := os.Open(configFilePath)
	if err != nil {
		log.Println(err)
		return
	}
	dec := json.NewDecoder(f)
	err = dec.Decode(&conf)
	if err != nil {
		log.Println(err)
		return
	}

	keyBytes, err := hex.DecodeString(conf.EncryptKey)
	if err != nil {
		log.Println("ENC_KEY is not hex decodable")
		return
	}

	key := [32]byte{}
	copy(key[:], keyBytes)

	readTemplate := func(lockTStore bool) error {
		tmpl := template.New("index.html")
		tmpl, err := tmpl.ParseFiles(conf.IndexTemplatePath)
		if lockTStore {
			tstore.Lock()
			defer tstore.Unlock()
		}
		tstore.tmpl = tmpl
		return err
	}

	err = readTemplate(false)
	if err != nil {
		log.Println(err)
		return
	}

	tstore.bgColor = bgColors[rand.Intn(len(bgColors))]

	var tapi *anaconda.TwitterApi
	if conf.TweetItems {
		tapi = anaconda.NewTwitterApiWithCredentials(conf.TwitterAccessToken, conf.TwitterAccessTokenSecret, conf.TwitterConsumerAPIKey, conf.TwitterConsumerSecretKey)
	}

	db, err := sql.Open("sqlite3", conf.AppDatabasePath)
	if err != nil {
		log.Println(err)
		return
	}

	defer db.Close()

	store := sim.NewStore()
	rate, err := limiter.NewRateFromFormatted(conf.RateLimit)
	if err != nil {
		log.Println(err)
		return
	}

	err = setupTables(db)
	if err != nil {
		log.Println(err)
		return
	}

	errs := updateItemsTable(db, addByColumn, addTextxColumn, addDescColumn, addTweetIDColumn, addEncLink, addEncDiscussLink)
	for _, err = range errs {
		log.Println(err)
	}

	rlMiddleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	go flow(tctx, db, conf, tapi, &key)

	go func() {
		sixMinTicker := time.NewTicker(6 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sixMinTicker.C:
				tctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
				defer cancel()
				flow(tctx, db, conf, tapi, &key)
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
				err := readTemplate(true)
				rand.Seed(time.Now().Unix())
				tstore.bgColor = bgColors[rand.Intn(len(bgColors))]
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	fetchItems := func(since time.Time) ([]*item, error) {
		var ids []int
		func() {
			tstore.RLock()
			defer tstore.RUnlock()
			ids = tstore.currentTop30ItemIds
		}()

		return fetchCurrentItems(db, since, ids)
	}

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(conf.StaticResourcesDirectoryPath))))

	r.Handle("/", rlMiddleware.Handler(withRequestHeadersLogging(func(w http.ResponseWriter, r *http.Request) {
		t, err := strconv.Atoi(r.URL.Query().Get("t"))
		var since time.Time
		if err != nil || t <= 8 {
			since = time.Now().Add(-1 * eightHrs)
		} else if t > 8 && t <= 16 {
			since = time.Now().Add(-2 * eightHrs)
		} else {
			since = time.Now().Add(-3 * eightHrs)
		}

		items, err := fetchItems(since)
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
				data := make(map[string]interface{})
				data["bgColor"] = tstore.bgColor
				data["items"] = items
				err = tstore.tmpl.Execute(w, data)
				if err != nil {
					log.Println(err)
				}
			}()
		}
	}))).Methods(http.MethodGet)

	r.Handle("/sitemap.xml", rlMiddleware.Handler(withRequestHeadersLogging(func(w http.ResponseWriter, r *http.Request) {
		eightHrsBack := time.Now().Add(-eightHrs)
		items, err := fetchItems(eightHrsBack)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}

		sm := sitemap.New()
		for _, it := range items {
			added := time.Unix(int64(it.Added), 0)
			h, err := encAndHex(it.URL, &key)
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
	}))).Methods(http.MethodGet)

	r.Handle("/feed/{type}", rlMiddleware.Handler(withRequestHeadersLogging(func(w http.ResponseWriter, r *http.Request) {
		eightHrsBack := time.Now().Add(-eightHrs)
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
	}))).Queries().Methods(http.MethodGet)

	r.Handle("/l/{hash}", rlMiddleware.Handler(withRequestHeadersLogging(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		h := vars["hash"]

		link, err := decFromHex(h, &key)
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
	}))).Methods(http.MethodGet, http.MethodPost)

	if conf.HaveRobotsTxt {
		r.Handle("/robots.txt", rlMiddleware.Handler(withRequestHeadersLogging(serveFile(conf.RobotsTextFilePath)))).Methods(http.MethodGet)
	}

	http.Handle("/", r)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", conf.HTTPPort),
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	log.Println(srv.ListenAndServe())
}

func serveFile(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

func withRequestHeadersLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Visiting : %s with method %s\n", r.URL.Path, r.Method)
		for h, v := range r.Header {
			log.Printf("%s : %s\n", h, strings.Join(v, " "))
		}
		ip := realIP(r)
		if ip != "" {
			log.Printf("Real IP : %s\n", ip)
		}

		next.ServeHTTP(w, r)
	}
}

func flow(ctx context.Context, db *sql.DB, conf *config, tapi *anaconda.TwitterApi, key *[32]byte) {
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

	items, err := fetchTopStories(ctx, 30)
	if err != nil {
		log.Println(err)
	}

	thirtyTwoHrsBack := time.Now().Add(-4 * eightHrs)

	err = updateItemsAddedTimeToNow(db, ids)
	if err != nil {
		log.Println(err)
	}

	idAndTweetIDs, err := selectItemsIDsBefore(db, thirtyTwoHrsBack.Unix())
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
		err = deleteItemsWith(db, olderItemsIDsNotInTop)
		if err != nil {
			log.Println(err)
		}

		if conf.TweetItems {
			errs := deleteTweets(ctx, tapi, tweetIDsFromOlderItemsToBeDeleted)
			for id, err := range errs {
				log.Printf("%d tweet deletion failed with %s\n", id, err)
			}
		}
	}

	eightHrsBack := time.Now().Add(-1 * eightHrs)
	itemsIDsFromLastEightHrs, err := selectItemsIDsAfter(db, eightHrsBack.Unix())
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

	updatedItems, err := fetchHNStoriesOf(ctx, olderItemsIDsInTop)
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

	existingItems, err := selectExistingPropsOfItemsByIDsAsc(db, itemIDs)
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

	for _, it := range items {
		if it.URL == "" {
			it.URL = fmt.Sprintf(hnPostLinkURL, it.ID)
		}
		if it.DiscussLink == "" {
			it.DiscussLink = fmt.Sprintf(hnPostLinkURL, it.ID)
		}
		if it.Domain == "" {
			domain, err := urlToDomain(it.URL)
			if err == nil {
				it.Domain = domain
			}
		}
		if it.EncryptedURL == "" {
			link := it.URL
			if link == "" {
				link = it.DiscussLink
			}
			h, _ := encAndHex(link, key)
			it.EncryptedURL = h
		}
		if it.EncryptedDiscussLink == "" {
			h, _ := encAndHex(it.DiscussLink, key)
			it.EncryptedDiscussLink = h
		}
	}

	if conf.TweetItems {
		errs := tweetItems(ctx, tapi, items)
		for id, err := range errs {
			log.Printf("%d tweeting failed with %s\n", id, err)
		}
	}

	err = insertOrReplaceItems(db, items)
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
