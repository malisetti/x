package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

func selectItemsBefore(db *sql.DB, t int64) ([]*item, error) {
	stmt := `SELECT id, link, added FROM items WHERE added >= %d`
	stmt = fmt.Sprintf(stmt, t)

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	return serializeRowsToItems(rows)
}

func selectItemsByIDs(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, link, added FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `)`

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	return serializeRowsToItems(rows)
}

func insertOrReplaceItems(db *sql.DB, items []*item) (sql.Result, error) {
	valueArgs := make([]string, 0, len(items)*3)
	valueArgsTmpl := "(%d, \"%s\", %s)"
	now := time.Now().Unix()
	for _, it := range items {
		if it.URL == "" {
			it.URL = fmt.Sprintf(hnPostLinkURL, it.ID)
		}
		added := fmt.Sprintf("COALESCE((SELECT added FROM items WHERE id = %d), %d)", it.ID, now)

		v := fmt.Sprintf(valueArgsTmpl, it.ID, it.URL, added)
		valueArgs = append(valueArgs, v)
	}
	stmt := fmt.Sprintf(`INSERT OR REPLACE INTO items (id, link, added) VALUES %s`, strings.Join(valueArgs, ","))

	return db.Exec(stmt)
}

func serializeRowsToItems(rows *sql.Rows) ([]*item, error) {
	defer rows.Close()
	var items []*item
	for rows.Next() {
		var it item
		err := rows.Scan(&it.ID, &it.URL, &it.Added)
		if err != nil {
			log.Println(err)
			continue
		}

		items = append(items, &it)
	}

	return items, nil
}
