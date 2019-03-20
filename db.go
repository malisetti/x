package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"
)

const addDescColumn string = "ALTER TABLE items ADD COLUMN `description` TEXT;"
const addTweetIDColumn string = "ALTER TABLE items ADD COLUMN `tweetID` INTEGER;"
const addByColumn string = "ALTER TABLE items ADD COLUMN `by` TEXT;"
const addTextxColumn string = "ALTER TABLE items ADD COLUMN `textx` TEXT;"
const addEncLink string = "ALTER TABLE items ADD COLUMN `encLink` TEXT;"
const addEncDiscussLink string = "ALTER TABLE items ADD COLUMN `encDiscussLink` TEXT;"

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

func selectItemsIDsAfter(db *sql.DB, t int64) ([]int, error) {
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
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx, encLink, encDiscussLink FROM items WHERE added >= %d`
	stmt = fmt.Sprintf(stmt, t)

	return execStmtAndGetItems(db, stmt)
}

func selectItemsIDsBefore(db *sql.DB, t int64) (map[int]int64, error) {
	stmt := `SELECT id, tweetID FROM items WHERE added <= %d`
	stmt = fmt.Sprintf(stmt, t)

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	ids := make(map[int]int64)
	for rows.Next() {
		var id int
		var tweetID int64
		var nullInt64 sql.NullInt64
		err := rows.Scan(&id, &nullInt64)
		if err != nil {
			log.Println(err)
			continue
		}

		if nullInt64.Valid {
			tweetID = nullInt64.Int64
		}

		ids[id] = tweetID
	}

	return ids, nil
}

func selectItemsBefore(db *sql.DB, t int64) ([]*item, error) {
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx, encLink, encDiscussLink FROM items WHERE added <= %d`
	stmt = fmt.Sprintf(stmt, t)

	return execStmtAndGetItems(db, stmt)
}

func selectExistingPropsOfItemsByIDsAsc(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, link, discussLink, domain, description, tweetID, encLink, encDiscussLink FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id ASC`

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var items []*item
	for rows.Next() {
		var it item
		var description, encryptedURL, encryptedDiscussLink sql.NullString
		var tweetID sql.NullInt64
		err := rows.Scan(&it.ID,
			&it.URL, &it.DiscussLink,
			&it.Domain, &description, &tweetID,
			&encryptedURL, &encryptedDiscussLink,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		it.Description = description.String

		if tweetID.Valid {
			it.TweetID = tweetID.Int64
		}

		it.EncryptedURL = encryptedURL.String
		it.EncryptedDiscussLink = encryptedDiscussLink.String

		items = append(items, &it)
	}

	return items, nil
}

func selectItemsByIDsAsc(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx, encLink, encDiscussLink FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id ASC`

	return execStmtAndGetItems(db, stmt)
}

func selectItemsByIDsDesc(db *sql.DB, ids []int) ([]*item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx, encLink, encDiscussLink FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id DESC`

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
		var description, by, textx, encryptedURL, encryptedDiscussLink sql.NullString
		var tweetID sql.NullInt64
		err := rows.Scan(&it.ID, &it.Title,
			&it.URL, &deleted,
			&dead, &it.DiscussLink,
			&it.Added, &it.Domain,
			&description, &tweetID,
			&by, &textx,
			&encryptedURL, &encryptedDiscussLink,
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

		it.Description = description.String

		if tweetID.Valid {
			it.TweetID = tweetID.Int64
		}
		it.By = by.String
		it.Textx = textx.String
		it.EncryptedURL = encryptedURL.String
		it.EncryptedDiscussLink = encryptedDiscussLink.String

		items = append(items, &it)
	}

	return items, nil
}

func updateItemsAddedTimeToNow(db *sql.DB, ids []int) error {
	now := time.Now().Unix()
	var args []interface{}
	args = append(args, now)

	var placeHolder []string
	for _, id := range ids {
		placeHolder = append(placeHolder, "?")
		args = append(args, id)
	}

	sqlStr := `UPDATE items SET added = ? WHERE id IN (` + strings.Join(placeHolder, ",") + `)`

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(sqlStr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)

	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func insertOrReplaceItems(db *sql.DB, items []*item) error {
	// id, title, url, deleted, dead, discussLink, added, domain, description, tweetID, by, textx, encLink, encDiscussLink
	sqlStr := "INSERT OR REPLACE INTO items (id, title, link, deleted, dead, discussLink, added, domain, description, tweetID, by, textx, encLink, encDiscussLink) VALUES (?, ?, ?, ?, ?, ?, COALESCE((SELECT added FROM items WHERE id = ?), ?), ?, ?, ?, ?, ?, ?, ?)"

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	stmt, err := tx.Prepare(sqlStr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Unix()
	for _, it := range items {
		deleted := 0
		dead := 0
		if it.Deleted {
			deleted = 1
		}
		if it.Dead {
			dead = 1
		}

		_, err := stmt.Exec(it.ID, it.Title, it.URL, deleted, dead, it.DiscussLink, it.ID, now, it.Domain, it.Description, it.TweetID, it.By, it.Textx, it.EncryptedURL, it.EncryptedDiscussLink)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
