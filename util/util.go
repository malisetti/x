// Package util provides utitlity functions.
package util

import (
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/mseshachalam/x/app"
)

const (
	headerXForwardedFor = "X-Forwarded-For"
	headerXRealIP       = "X-Real-IP"
)

// IntsToChan converts array of ints to chan of int
func IntsToChan(ints []int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, i := range ints {
			out <- i
		}
	}()
	return out
}

// Int64sToChan converts array of ints to chan of int
func Int64sToChan(ints []int64) <-chan int64 {
	out := make(chan int64)
	go func() {
		defer close(out)
		for _, i := range ints {
			out <- i
		}
	}()
	return out
}

// ItemsToChan converts array of items to chan of item
func ItemsToChan(items []*app.Item) <-chan *app.Item {
	itemsChan := make(chan *app.Item)
	go func() {
		defer close(itemsChan)
		for _, it := range items {
			itemsChan <- it
		}
	}()
	return itemsChan
}

// URLToDomain extracts domain from given link
func URLToDomain(link string) (string, error) {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	parts := strings.Split(u.Hostname(), ".")
	if len(parts[0]) > 4 {
		return strings.Join(parts, "."), nil
	}
	if strings.HasPrefix(parts[0], "www") {
		return strings.Join(parts[1:], "."), nil
	}

	return strings.Join(parts, "."), nil
}

// RealIP tries to extract real ip from request r using X-Forwarded-For and X-Real-IP headers
func RealIP(r *http.Request) string {
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
