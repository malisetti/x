package server

import (
	"log"
	"net/http"
	"strings"

	"github.com/mseshachalam/x/util"
)

// RequiredHeaders are the headers we need to care about
var RequiredHeaders = []string{"Cf-Ipcountry", "Cf-Connecting-Ip", "X-Forwarded-For"}

// CrawlerAliases are the names of possible crawler's user agents
var CrawlerAliases = []string{"bot", "crawler", "spider", "trendsmapresolver", "fetcher"}

// WithRequestHeadersLogging logs headers from RequiredHeaders
func WithRequestHeadersLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headers := []string{}
		for _, h := range RequiredHeaders {
			v := r.Header.Get(h)
			if v == "" {
				continue
			}
			headers = append(headers, h, v)
		}
		ip := util.RealIP(r)
		if ip != "" && len(headers) > 0 {
			headers = append(headers, "Real-IP", ip)
		}
		if len(headers) > 0 {
			log.Println(strings.Join(headers, " - "))
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
				log.Println("request rejected with forbidden status")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	}
}
