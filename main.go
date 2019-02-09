package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	fmt.Println("hello world")

	db, err := sql.Open("sqlite3", "./app.db")
	if err != nil {
		log.Println(err)
		return
	}

	// stmt, err := db.Prepare("INSERT INTO items(id, link, added) values(?,?,?)")
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	items, err := fetchTopStories(ctx, 30)
	if err != nil {
		log.Println(err)
		return
	}

	valueArgs := make([]string, 0, len(items)*3)
	valueArgsTmpl := "(%d, \"%s\", %s)"
	for _, it := range items {
		if it.URL == "" {
			it.URL = fmt.Sprintf(hnPostLinkURL, it.ID)
		}
		now := time.Now().Unix()
		added := fmt.Sprintf("COALESCE((SELECT added FROM items WHERE id = %d), %d)", it.ID, now)

		v := fmt.Sprintf(valueArgsTmpl, it.ID, it.URL, added)
		valueArgs = append(valueArgs, v)
	}
	stmt := fmt.Sprintf(`INSERT OR REPLACE INTO items (id, link, added) VALUES %s`, strings.Join(valueArgs, ","))

	log.Println(stmt)

	res, err := db.Exec(stmt)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(res)
}
