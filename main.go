package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"os"

	"net"
	"net/http"
	"strings"
	"time"

	"sync"

	"github.com/gorilla/mux"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sim "github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/ChimeraCoder/anaconda"
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

	middleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))

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

	const headerXForwardedFor = "X-Forwarded-For"
	const headerXRealIP = "X-Real-IP"
	realIP := func(r *http.Request) string {
		ra := r.RemoteAddr
		if ip := r.Header.Get(headerXForwardedFor); ip != "" {
			ra = strings.Split(ip, ", ")[0]
		} else if ip := r.Header.Get(headerXRealIP); ip != "" {
			ra = ip
		} else {
			ra, _, _ = net.SplitHostPort(ra)
		}

		return ra
	}

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(os.Getenv("STATIC_DIR")))))

	r.Handle("/", middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// log CF- headers
		for h, v := range r.Header {
			h = strings.ToUpper(h) // headers are case insensitive
			if strings.HasPrefix(h, "CF-") {
				log.Printf("%s : %s\n", strings.Replace(h, "CF-", "", -1), strings.Join(v, " "))
			}
		}
		log.Println(realIP(r))
		log.Println(r.UserAgent())

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

		func() {
			tstore.RLock()
			defer tstore.RUnlock()
			err = tstore.tmpl.Execute(w, items)
			if err != nil {
				log.Println(err)
			}
		}()
	}))).Methods(http.MethodHead, http.MethodGet)

	r.Handle("/json", middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			log.Println(err)
		}
	}))).Methods(http.MethodGet)

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
			tweetIDsFromOlderItemsToBeDeleted = append(tweetIDsFromOlderItemsToBeDeleted, it.TweetID)
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
