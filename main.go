package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"

	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	sim "github.com/ulule/limiter/v3/drivers/store/memory"

	_ "github.com/mattn/go-sqlite3"
)

const (
	eightHrs = 8 * 60 * 60 * time.Second
)

func main() {
	fmt.Println("hello world")

	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Println(err)
		return
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	items, err := fetchTopStories(ctx, 30)
	if err != nil {
		log.Println(err)
		return
	}
	res, err := insertOrReplaceItems(db, items)
	if err != nil {
		log.Println(err)
		return
	}

	items, err = selectItemsByIDs(db, []int{19113147, 19115964})
	b, _ := json.MarshalIndent(items, "", "  ")
	log.Println(string(b), err)

	log.Println(res.RowsAffected())

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

	store := sim.NewStore()
	// Define a limit rate to 5 requests per minute.
	rate, err := limiter.NewRateFromFormatted("5-M")
	if err != nil {
		panic(err)
	}

	middleware := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))

	tmpl := template.New("index.html")
	tmpl, err = tmpl.ParseFiles("./index.html")
	if err != nil {
		log.Println(err)
		return
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

		eightHrsBack := time.Now().Add(-eightHrs)

		log.Println(eightHrsBack.Unix())

		items, err := selectItemsBefore(db, eightHrsBack.Unix())
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

	srv := &http.Server{Addr: fmt.Sprintf(":%d", 8080)}

	log.Println(srv.ListenAndServe())
}
