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

	valueStrings := make([]string, 0, len(items))
	valueArgs := make([]interface{}, 0, len(items)*3)
	for _, it := range items {
		if it.URL == "" {
			it.URL = fmt.Sprintf(hnPostLinkURL, it.ID)
		}
		valueStrings = append(valueStrings, "(?, ?, ?)")
		valueArgs = append(valueArgs, it.ID)
		valueArgs = append(valueArgs, it.URL)
		now := time.Now().Unix()
		valueArgs = append(valueArgs, now)
	}
	stmt := fmt.Sprintf("INSERT INTO items (id, link, added) VALUES %s", strings.Join(valueStrings, ","))
	res, err := db.Exec(stmt, valueArgs...)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println(res)
}
