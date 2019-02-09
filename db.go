package main

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

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
