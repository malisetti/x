package main

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"os"

	"net"
	"net/http"
	"strings"
	"time"

	"sync"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sim "github.com/ulule/limiter/v3/drivers/store/memory"

	_ "github.com/mattn/go-sqlite3"
)

const (
	port     = "8080"
	eightHrs = 8 * 60 * 60 * time.Second
)

type tempStore struct {
	sync.RWMutex
	currentTop30ItemIds []int
}

var tstore tempStore

func main() {
	indexTemplatePath := os.Getenv("INDEX_TMPL_PATH")
	tmpl := template.New("index.html")
	tmpl, err := tmpl.ParseFiles(indexTemplatePath)
	if err != nil {
		log.Println(err)
		return
	}

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

	middleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	flow(ctx, db)

	go func() {
		fiveMinTicker := time.NewTicker(5 * time.Minute)
		for {
			select {
			case <-ctx.Done():
				return
			case <-fiveMinTicker.C:
				flow(ctx, db)
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

	http.Handle("/", middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		eightHrsBack := time.Now().Add(-eightHrs)
		oIds, err := selectItemsIdsBefore(db, eightHrsBack.Unix())
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}
		currentTopPlusEightHrs := append(ids, oIds...)

		items, err := selectItemsByIDs(db, currentTopPlusEightHrs)
		if err != nil {
			fmt.Fprintf(w, "%s", err)
			return
		}

		data := make(map[int]*item)
		for _, it := range items {
			data[it.ID] = it
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Println(err)
		}
	})))

	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = port
	}
	srv := &http.Server{Addr: fmt.Sprintf(":%s", httpPort)}

	log.Println(srv.ListenAndServe())
}

func flow(ctx context.Context, db *sql.DB) {
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
		}
	}
	if len(olderItemsIDsNotInTop) > 0 {
		err = deleteItemsWith(db, olderItemsIDsNotInTop)
		if err != nil {
			log.Println(err)
		}
	}

	_, err = insertOrReplaceItems(db, items)
	if err != nil {
		log.Println(err)
	}
}
