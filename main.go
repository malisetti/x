// This application keeps any article that appeared on hacker news (http://news.ycombinator.com/) for eight hours after they left the homepage.
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

	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sim "github.com/ulule/limiter/v3/drivers/store/memory"

	"github.com/ChimeraCoder/anaconda"
	_ "github.com/mattn/go-sqlite3"

	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/dbp"
	"github.com/mseshachalam/x/flow"
	"github.com/mseshachalam/x/hn"
	"github.com/mseshachalam/x/server"
)

func main() {
	var tstore app.TempStore
	var conf *app.Config
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
		tstore.Tmpl = tmpl
		return err
	}

	err = readTemplate(false)
	if err != nil {
		log.Println(err)
		return
	}

	tstore.BgColor = app.BgColors[rand.Intn(len(app.BgColors))]

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

	err = dbp.SetupTables(db)
	if err != nil {
		log.Println(err)
		return
	}

	errs := dbp.UpdateItemsTable(db, dbp.AddByColumn, dbp.AddTextxColumn, dbp.AddDescColumn, dbp.AddTweetIDColumn, dbp.AddEncLink, dbp.AddEncDiscussLink)
	for _, err = range errs {
		log.Println(err)
	}

	rlMiddleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	tctx, tcancel := context.WithTimeout(ctx, 5*time.Minute)
	defer tcancel()

	go flow.Flow(tctx, &tstore, db, conf, tapi, &key)

	go func() {
		sixMinTicker := time.NewTicker(6 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				return
			case <-sixMinTicker.C:
				tctx, tcancel := context.WithTimeout(ctx, 5*time.Minute)
				defer tcancel()
				flow.Flow(tctx, &tstore, db, conf, tapi, &key)
			}
		}
	}()

	go func() {
		eightHrsTicker := time.NewTicker(app.EightHrs)
		for {
			select {
			case <-ctx.Done():
				return
			case <-eightHrsTicker.C:
				err := readTemplate(true)
				rand.Seed(time.Now().Unix())
				tstore.BgColor = app.BgColors[rand.Intn(len(app.BgColors))]
				if err != nil {
					log.Println(err)
				}
			}
		}
	}()

	fetchItems := func(since time.Time) ([]*app.Item, error) {
		var ids []int
		func() {
			tstore.RLock()
			defer tstore.RUnlock()
			ids = tstore.CurrentTop30ItemIds
		}()

		return hn.FetchCurrentItems(db, since, ids)
	}

	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/",
			http.FileServer(http.Dir(conf.StaticResourcesDirectoryPath))))

	r.Handle("/", rlMiddleware.Handler(server.WithRequestHeadersLogging(server.RootHandler(fetchItems, &tstore)))).Methods(http.MethodGet)

	r.Handle("/sitemap.xml", rlMiddleware.Handler(server.WithRequestHeadersLogging(server.SitemapHandler(fetchItems, &key)))).Methods(http.MethodGet)

	r.Handle("/feed/{type}", rlMiddleware.Handler(server.WithRequestHeadersLogging(server.FeedHandler(fetchItems)))).Methods(http.MethodGet)

	r.Handle("/l/{hash}", rlMiddleware.Handler(server.WithRequestHeadersLogging(server.LinkHandler(&key)))).Methods(http.MethodGet, http.MethodPost)

	if conf.HaveRobotsTxt {
		r.Handle("/robots.txt", rlMiddleware.Handler(server.WithRequestHeadersLogging(server.FileHandler(conf.RobotsTextFilePath)))).Methods(http.MethodGet)
	}

	http.Handle("/", r)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", conf.HTTPPort),
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	log.Println(srv.ListenAndServe())
}
