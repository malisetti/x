// Package server provides http middlewares.
package server

import (
	"log"
	"net/http"

	"github.com/mseshachalam/x/util"
)

// RequiredHeaders are the headers we need to care about
var RequiredHeaders = []string{"User-Agent", "Cf-Ipcountry", "Accept", "Cf-Connecting-Ip", "X-Forwarded-For"}

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
