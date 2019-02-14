package main

import (
	"database/sql"
	"fmt"
	"html"
	"log"
	"strings"
	"time"
)

const addDescColumn string = "ALTER TABLE items ADD COLUMN `description` TEXT;"
const addTweetIDColumn string = "ALTER TABLE items ADD COLUMN `tweetID` INTEGER"
const addByColumn string = "ALTER TABLE items ADD COLUMN `by` TEXT;"
const addTextxColumn string = "ALTER TABLE items ADD COLUMN `textx` TEXT;"

func setupTables(db *sql.DB) error {
	stmt := "CREATE TABLE IF NOT EXISTS `items` (`id`	INTEGER PRIMARY KEY AUTOINCREMENT,`link`	TEXT NOT NULL,`added`	INTEGER NOT NULL,`title`	TEXT,`deleted`	INTEGER,`dead`	INTEGER,`discussLink`	TEXT,`domain`	TEXT)"

	_, err := db.Exec(stmt)
	return err
}

func updateItemsTable(db *sql.DB, stmts ...string) []error {
	var errs []error
	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func deleteItemsWith(db *sql.DB, ids []int) error {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}

	stmt := `DELETE FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `)`
	stmt = fmt.Sprintf(stmt)
	_, err := db.Exec(stmt)
	return err
}

func deleteOlderItems(db *sql.DB, t int64) error {
	stmt := `DELETE FROM items WHERE added < %d`
	stmt = fmt.Sprintf(stmt, t)
	_, err := db.Exec(stmt)
	return err
}

func selectItemsIdsBefore(db *sql.DB, t int64) ([]int, error) {
	stmt := `SELECT id FROM items WHERE added >= %d`
	stmt = fmt.Sprintf(stmt, t)

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var ids []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err)
			continue
		}

		ids = append(ids, id)
	}

	return ids, nil
}

func selectItemsAfter(db *sql.DB, t int64) ([]*item, error) {
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx FROM items WHERE added >= %d`
	stmt = fmt.Sprintf(stmt, t)

	return execStmtAndGetItems(db, stmt)
}

func selectItemsBefore(db *sql.DB, t int64) ([]*item, error) {
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx FROM items WHERE added <= %d`
	stmt = fmt.Sprintf(stmt, t)

	return execStmtAndGetItems(db, stmt)
}

func selectItemsByIDsAsc(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id ASC`

	return execStmtAndGetItems(db, stmt)
}

func selectItemsByIDsDesc(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id DESC`

	return execStmtAndGetItems(db, stmt)
}

func execStmtAndGetItems(db *sql.DB, stmt string) ([]*item, error) {
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var items []*item
	for rows.Next() {
		var it item
		var dead, deleted int
		var description, by, textx sql.NullString
		var tweetID sql.NullInt64
		err := rows.Scan(&it.ID, &it.Title,
			&it.URL, &deleted,
			&dead, &it.DiscussLink,
			&it.Added, &it.Domain,
			&description, &tweetID,
			&by, &textx,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		if dead == 1 {
			it.Dead = true
		}
		if deleted == 1 {
			it.Deleted = true
		}

		it.Descriprion = description.String

		if tweetID.Valid {
			it.TweetID = tweetID.Int64
		}
		it.By = by.String
		it.Textx = by.String

		items = append(items, &it)
	}

	return items, nil
}

func insertOrReplaceItems(db *sql.DB, items []*item) (sql.Result, error) {
	var valueArgs []string

	// id, title, url, deleted, dead, discussLink, added, domain, description, tweetID, by, textx
	valueArgsTmpl := "(%d, \"%s\", \"%s\", %d, %d, \"%s\", %s, \"%s\", \"%s\", %d, \"%s\", \"%s\")"
	now := time.Now().Unix()
	for _, it := range items {
		added := fmt.Sprintf("COALESCE((SELECT added FROM items WHERE id = %d), %d)", it.ID, now)

		var deleted int
		if it.Deleted {
			deleted = 1
		}
		var dead int
		if it.Dead {
			dead = 1
		}

		textx := html.EscapeString(it.Textx)
		description := html.EscapeString(it.Descriprion)

		v := fmt.Sprintf(valueArgsTmpl, it.ID, it.Title, it.URL, deleted, dead, it.DiscussLink, added, it.Domain, description, it.TweetID, it.By, textx)
		valueArgs = append(valueArgs, v)
	}
	stmt := fmt.Sprintf(`INSERT OR REPLACE INTO items (id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx) VALUES %s`, strings.Join(valueArgs, ","))

	return db.Exec(stmt)
}
