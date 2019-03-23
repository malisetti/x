package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/mseshachalam/x/util"
)

// RequiredHeaders are the headers we need to care about
var RequiredHeaders = []string{"User-Agent", "Cf-Ipcountry", "Accept", "Cf-Connecting-Ip", "X-Forwarded-For"}

// CrawlerAliases are the names of possible crawler's user agents
var CrawlerAliases = []string{"bot", "crawler", "spider", "trendsmapresolver", "fetcher"}

// WithRequestHeadersLogging logs headers from RequiredHeaders
func WithRequestHeadersLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Visiting : %s with method : %s\n", r.URL.Path, r.Method)
		for _, h := range RequiredHeaders {
			log.Printf("%s : %s\n", h, r.Header.Get(h))
		}
		ip := util.RealIP(r)
		if ip != "" {
			log.Printf("Real IP : %s\n", ip)
		}

		next.ServeHTTP(w, r)
	}
}

// WithBotsAndCrawlersBlocking blocks bots and crawlers based on user agents of requests
func WithBotsAndCrawlersBlocking(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ua := r.Header.Get("User-Agent")
		ua = strings.ToLower(ua)

		for _, ca := range CrawlerAliases {
			if strings.Contains(ua, ca) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}
