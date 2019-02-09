package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

func setupTables(db *sql.DB) error {
	stmt := "CREATE TABLE IF NOT EXISTS `items` (`id`	INTEGER PRIMARY KEY AUTOINCREMENT,`link`	TEXT NOT NULL,`added`	INTEGER NOT NULL,`title`	TEXT,`deleted`	INTEGER,`dead`	INTEGER,`discussLink`	TEXT,`domain`	TEXT)"

	_, err := db.Exec(stmt)
	return err
}

func selectItemsBefore(db *sql.DB, t int64) ([]*item, error) {
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain FROM items WHERE added >= %d`
	stmt = fmt.Sprintf(stmt, t)

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var items []*item
	for rows.Next() {
		var it item
		err := rows.Scan(&it.ID, &it.Title,
			&it.URL, &it.Deleted,
			&it.Dead, &it.DiscussLink,
			&it.Added, &it.Domain)
		if err != nil {
			log.Println(err)
			continue
		}

		items = append(items, &it)
	}

	return items, nil
}

func selectItemsByIDs(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `)`

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var items []*item
	for rows.Next() {
		var it item
		err := rows.Scan(&it.ID, &it.Title,
			&it.URL, &it.Deleted,
			&it.Dead, &it.DiscussLink,
			&it.Added, &it.Domain)
		if err != nil {
			log.Println(err)
			continue
		}

		items = append(items, &it)
	}

	return items, nil
}

func insertOrReplaceItems(db *sql.DB, items []*item) (sql.Result, error) {
	valueArgs := make([]string, 0, len(items)*3)

	// id, title, url, deleted, dead, discussLink, added, domain
	valueArgsTmpl := "(%d, \"%s\", \"%s\", %d, %d, \"%s\", %s, \"%s\")"
	now := time.Now().Unix()
	for _, it := range items {
		if it.URL == "" {
			it.URL = fmt.Sprintf(hnPostLinkURL, it.ID)
		}
		added := fmt.Sprintf("COALESCE((SELECT added FROM items WHERE id = %d), %d)", it.ID, now)
		discussLink := fmt.Sprintf(hnPostLinkURL, it.ID)
		domain, err := urlToDomain(it.URL)
		if err != nil {
			log.Println(err)
		}

		var deleted int
		if it.Deleted {
			deleted = 1
		}
		var dead int
		if it.Dead {
			dead = 1
		}

		v := fmt.Sprintf(valueArgsTmpl, it.ID, it.Title, it.URL, deleted, dead, discussLink, added, domain)
		valueArgs = append(valueArgs, v)
	}
	stmt := fmt.Sprintf(`INSERT OR REPLACE INTO items (id, title, link, deleted, dead, discussLink, added, domain) VALUES %s`, strings.Join(valueArgs, ","))

	return db.Exec(stmt)
}
