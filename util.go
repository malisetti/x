package main

import (
	"net"
	"net/http"
	"net/url"
	"strings"
)

const (
	headerXForwardedFor = "X-Forwarded-For"
	headerXRealIP       = "X-Real-IP"
)

var r = strings.NewReplacer("http://", "", "https://", "", "www.", "", "www2.", "", "www3.", "")

func intsToChan(itemIds []int) <-chan int {
	ids := make(chan int)
	go func() {
		defer close(ids)
		for _, itID := range itemIds {
			ids <- itID
		}
	}()
	return ids
}

func int64sToChan(itemIds []int64) <-chan int64 {
	ids := make(chan int64)
	go func() {
		defer close(ids)
		for _, itID := range itemIds {
			ids <- itID
		}
	}()
	return ids
}

func itemsToChan(items []*item) <-chan *item {
	itemsChan := make(chan *item)
	go func() {
		defer close(itemsChan)
		for _, it := range items {
			itemsChan <- it
		}
	}()
	return itemsChan
}

func urlToDomain(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	parts := strings.Split(u.Hostname(), ".")
	if len(parts) >= 2 {
		domain := parts[len(parts)-2] + "." + parts[len(parts)-1]
		return domain, nil
	}

	return r.Replace(u.Hostname()), nil
}

func realIP(r *http.Request) string {
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
