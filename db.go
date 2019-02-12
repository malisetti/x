package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

const addDescColumn string = "ALTER TABLE items ADD COLUMN `description` text;"
const addImgsColumn string = "ALTER TABLE items ADD COLUMN `images` text;"

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
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, images FROM items WHERE added >= %d`
	stmt = fmt.Sprintf(stmt, t)

	return execStmtAndGetItems(db, stmt)
}

func selectItemsBefore(db *sql.DB, t int64) ([]*item, error) {
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, images FROM items WHERE added <= %d`
	stmt = fmt.Sprintf(stmt, t)

	return execStmtAndGetItems(db, stmt)
}

func selectItemsByIDsAsc(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, images FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id ASC`

	return execStmtAndGetItems(db, stmt)
}

func selectItemsByIDsDesc(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, images FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id DESC`

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
		var description, images sql.NullString
		err := rows.Scan(&it.ID, &it.Title,
			&it.URL, &deleted,
			&dead, &it.DiscussLink,
			&it.Added, &it.Domain,
			&description, &images)
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
		if images.String != "" {
			it.Images = strings.Split(images.String, "|")
		}

		items = append(items, &it)
	}

	return items, nil
}

func insertOrReplaceItems(db *sql.DB, items []*item) (sql.Result, error) {
	var valueArgs []string

	// id, title, url, deleted, dead, discussLink, added, domain, description, images
	valueArgsTmpl := "(%d, \"%s\", \"%s\", %d, %d, \"%s\", %s, \"%s\", \"%s\", \"%s\")"
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

		var images string
		if len(it.Images) > 0 {
			images = strings.Join(it.Images, "|")
		}

		v := fmt.Sprintf(valueArgsTmpl, it.ID, it.Title, it.URL, deleted, dead, discussLink, added, domain, it.Descriprion, images)
		valueArgs = append(valueArgs, v)
	}
	stmt := fmt.Sprintf(`INSERT OR REPLACE INTO items (id, title, link, deleted, dead, discussLink, added, domain, description, images) VALUES %s`, strings.Join(valueArgs, ","))

	return db.Exec(stmt)
}
