// This application keeps any article that appeared on hacker news (http://news.ycombinator.com/) for eight hours after they left the homepage.
package main

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/acme/autocert"

	"github.com/gorilla/mux"
	apachelog "github.com/lestrrat-go/apache-logformat"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sim "github.com/ulule/limiter/v3/drivers/store/memory"

	_ "github.com/mattn/go-sqlite3"

	"github.com/mseshachalam/x/app"
	"github.com/mseshachalam/x/dbp"
	"github.com/mseshachalam/x/hn"
	"github.com/mseshachalam/x/server"
)

func main() {
	var tstore app.TempStore
	var conf *app.Config
	configFilePath := os.Getenv("APP_CONFIG_PATH")
	f, err := os.Open(configFilePath)
	if err != nil {
		log.Printf("unable to open '%s' given by '%s' env: failed with error %s", configFilePath, "APP_CONFIG_PATH", err.Error())
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

	errs := dbp.UpdateItemsTable(db, dbp.AddByColumn, dbp.AddTextxColumn, dbp.AddDescColumn, dbp.AddEncLink, dbp.AddEncDiscussLink)
	for _, err = range errs {
		log.Println(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	hnMaintainer := &hn.HackerNewsMaintainer{
		PediodicBringer: &hn.HackerNewsPeriodicBringer{
			Ctx:      ctx,
			Interval: 5 * time.Minute,
		},
		Storage: db,
		Ctx:     ctx,
		Key:     &key,
	}

	go hnMaintainer.Maintain()

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

	allowedMethods := []string{http.MethodGet}
	if conf.EnableCors {
		allowedMethods = append(allowedMethods, http.MethodOptions)
	}

	rlMiddleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))

	handlers := &server.Server{
		Storage:    db,
		FetchItems: fetchItems,
		TStore:     &tstore,
	}

	r.Handle("/json", rlMiddleware.Handler(server.WithRequestHeadersLogging(handlers.JSONHandler(conf.EnableCors)))).Methods(allowedMethods...)

	r.Handle("/classic", rlMiddleware.Handler(server.WithRequestHeadersLogging(handlers.HTMLHandler()))).Methods(http.MethodGet)

	r.Handle("/sitemap.xml", rlMiddleware.Handler(server.WithRequestHeadersLogging(handlers.SitemapHandler(&key)))).Methods(http.MethodGet)

	r.Handle("/feed/{type}", rlMiddleware.Handler(server.WithRequestHeadersLogging(handlers.FeedHandler()))).Methods(http.MethodGet)

	r.Handle("/l/{hash}", rlMiddleware.Handler(server.WithRequestHeadersLogging(server.WithBotsAndCrawlersBlocking(handlers.LinkHandler(&key))))).Methods(http.MethodGet, http.MethodPost)

	if conf.HaveRobotsTxt {
		r.Handle("/robots.txt", rlMiddleware.Handler(server.WithRequestHeadersLogging(handlers.FileHandler(conf.RobotsTextFilePath)))).Methods(http.MethodGet)
	}

	r.PathPrefix("/").Handler(http.FileServer(http.Dir(conf.StaticResourcesDirectoryPath)))

	http.Handle("/", apachelog.CombinedLog.Wrap(r, os.Stderr))

	var wg sync.WaitGroup

	srv := &http.Server{
		Addr:           fmt.Sprintf(":%s", conf.HTTPPort),
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		IdleTimeout:    2 * time.Second,
		MaxHeaderBytes: (1 << 20) / 10,
		Handler:        r,
	}

	if conf.RunHTTPS {
		m := autocert.Manager{
			Email:      "abbiya@gmail.com",
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist("8hrs.xyz"),
			Cache:      autocert.DirCache(conf.AutoCertCacheDir),
		}

		srv.Addr = ":443"
		srv.TLSConfig = m.TLSConfig()

		go func() {
			wg.Add(1)
			defer wg.Done()
			err := srv.ListenAndServeTLS("", "")
			if err == http.ErrServerClosed {
				err = nil
			}
			if err != nil {
				fmt.Printf("Failed to start https, error: %s\n", err)
			} else {
				fmt.Printf("HTTPS server shutdown gracefully\n")
			}
		}()

		srv.Handler = m.HTTPHandler(srv.Handler)
	}

	go func() {
		wg.Add(1)
		defer wg.Done()
		err := srv.ListenAndServe()
		// mute error caused by Shutdown()
		if err == http.ErrServerClosed {
			err = nil
		}
		if err != nil {
			log.Printf("Failed to start http, error: %s\n", err)
		} else {
			log.Printf("HTTP server shutdown gracefully\n")
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, os.Interrupt, os.Kill)
	done := make(chan bool)
	go func() {
		for {
			select {
			case sig := <-c:
				log.Printf("Got signal %s\n", sig)
				if srv != nil {
					srv.Shutdown(ctx)
					cancel()
				}
			case <-done:
				return
			}
		}
	}()

	wg.Wait()
	done <- true
}
