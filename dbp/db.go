// Package dbp provides access to storage and querying.
package dbp

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mseshachalam/x/app"
)

const (
	// AddDescColumn adds description col
	AddDescColumn = "ALTER TABLE items ADD COLUMN `description` TEXT;"
	// AddByColumn adds by col
	AddByColumn = "ALTER TABLE items ADD COLUMN `by` TEXT;"
	// AddTextxColumn adds textx col
	AddTextxColumn = "ALTER TABLE items ADD COLUMN `textx` TEXT;"
	// AddEncLink adds encLink col
	AddEncLink = "ALTER TABLE items ADD COLUMN `encLink` TEXT;"
	// AddEncDiscussLink adds encDiscussLink col
	AddEncDiscussLink = "ALTER TABLE items ADD COLUMN `encDiscussLink` TEXT;"
)

// CreateTablesStmts contains needed sql stmts to setup required tables
var CreateTablesStmts = []string{
	"CREATE TABLE IF NOT EXISTS `items` (`id`	INTEGER PRIMARY KEY AUTOINCREMENT,`link`	TEXT NOT NULL,`added`	INTEGER NOT NULL,`title`	TEXT,`deleted`	INTEGER,`dead`	INTEGER,`discussLink`	TEXT,`domain`	TEXT)",
}

// SetupTables creates items table
func SetupTables(db *sql.DB, stmts []string) error {
	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateItemsTable executes stmts on db
func UpdateItemsTable(db *sql.DB, stmts ...string) map[string]error {
	errs := make(map[string]error)
	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		if err != nil {
			errs[stmt] = err
		}
	}

	return errs
}

// DeleteItemsWith deletes items with given ids
func DeleteItemsWith(db *sql.DB, ids []int) error {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}

	stmt := `DELETE FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `)`
	_, err := db.Exec(stmt)
	return err
}

// DeleteOlderItems deletes items that are olders than t
func DeleteOlderItems(db *sql.DB, t int64) error {
	stmt := `DELETE FROM items WHERE added < %d`
	stmt = fmt.Sprintf(stmt, t)
	_, err := db.Exec(stmt)
	return err
}

// SelectItemsIdsBefore selects items that are added after t, unix timestamp
func SelectItemsIdsBefore(db *sql.DB, t int64) ([]int, error) {
	stmt := `SELECT id FROM items WHERE added <= %d`
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

// SelectItemsIdsBeforeAndNotOf selects items that are added after t, unix timestamp and not of ids
func SelectItemsIdsBeforeAndNotOf(db *sql.DB, t int64, ids []int) ([]int, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id FROM items WHERE added <= %d AND id NOT IN (` + strings.Join(idsStr, ",") + `)`
	stmt = fmt.Sprintf(stmt, t)

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var oids []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err)
			continue
		}

		oids = append(oids, id)
	}

	return oids, nil
}

// SelectItemsIDsAfter selects items that are added before t
func SelectItemsIDsAfter(db *sql.DB, t int64) ([]int, error) {
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

// SelectItemsIDsAfterAndNotOf selects items that are added before t and not of ids
func SelectItemsIDsAfterAndNotOf(db *sql.DB, t int64, ids []int) ([]int, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id FROM items WHERE added >= %d AND id NOT IN (` + strings.Join(idsStr, ",") + `)`
	stmt = fmt.Sprintf(stmt, t)

	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var oids []int
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			log.Println(err)
			continue
		}

		oids = append(oids, id)
	}

	return oids, nil
}

// SelectItemsAfter selects items that are added after t
func SelectItemsAfter(db *sql.DB, t int64, order bool) ([]*app.Item, error) {
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, by, textx, encLink, encDiscussLink FROM items WHERE added >= %d`
	if order {
		stmt += " ORDER BY id ASC"
	} else {
		stmt += " ORDER BY id DESC"
	}
	stmt = fmt.Sprintf(stmt, t)

	return execStmtAndGetItems(db, stmt)
}

// SelectExistingPropsOfItemsByIDsAsc selects items details that are not from hn for given ids
func SelectExistingPropsOfItemsByIDsAsc(db *sql.DB, ids []int) ([]*app.Item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, link, discussLink, domain, description, encLink, encDiscussLink FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id ASC`
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var items []*app.Item
	for rows.Next() {
		var it app.Item
		var description, encryptedURL, encryptedDiscussLink sql.NullString
		err := rows.Scan(&it.ID,
			&it.URL, &it.DiscussLink,
			&it.Domain, &description,
			&encryptedURL, &encryptedDiscussLink,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		it.Description = description.String

		it.EncryptedURL = encryptedURL.String
		it.EncryptedDiscussLink = encryptedDiscussLink.String

		items = append(items, &it)
	}

	return items, nil
}

// SelectItemsByIDsAsc selects items for given ids in asc order
func SelectItemsByIDsAsc(db *sql.DB, ids []int) ([]*app.Item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, by, textx, encLink, encDiscussLink FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id ASC`

	return execStmtAndGetItems(db, stmt)
}

// SelectItemsByIDsDesc selects items for given ids in desc order
func SelectItemsByIDsDesc(db *sql.DB, ids []int) ([]*app.Item, error) {
	var idsStr []string
	for _, id := range ids {
		idsStr = append(idsStr, fmt.Sprintf("%d", id))
	}
	stmt := `SELECT id, title, link, deleted, dead, discussLink, added, domain, description, by, textx, encLink, encDiscussLink FROM items WHERE id IN (` + strings.Join(idsStr, ",") + `) ORDER BY id DESC`

	return execStmtAndGetItems(db, stmt)
}

func execStmtAndGetItems(db *sql.DB, stmt string) ([]*app.Item, error) {
	rows, err := db.Query(stmt)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	var items []*app.Item
	for rows.Next() {
		var it app.Item
		var dead, deleted int
		var description, by, textx, encryptedURL, encryptedDiscussLink sql.NullString
		err := rows.Scan(&it.ID, &it.Title,
			&it.URL, &deleted,
			&dead, &it.DiscussLink,
			&it.Added, &it.Domain,
			&description,
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
		it.By = by.String
		it.EncryptedURL = encryptedURL.String
		it.EncryptedDiscussLink = encryptedDiscussLink.String

		items = append(items, &it)
	}

	return items, nil
}

// UpdateItemsAddedTimeToNow updates added time for given ids
func UpdateItemsAddedTimeToNow(db *sql.DB, ids []int) error {
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

// InsertOrReplaceItems inserts or replaces given items
func InsertOrReplaceItems(db *sql.DB, items []*app.Item) error {
	// id, title, url, deleted, dead, discussLink, added, domain, description, by, textx, encLink, encDiscussLink
	sqlStr := "INSERT OR REPLACE INTO items (id, title, link, deleted, dead, discussLink, added, domain, description, by, textx, encLink, encDiscussLink) VALUES (?, ?, ?, ?, ?, ?, COALESCE((SELECT added FROM items WHERE id = ?), ?), ?, ?, ?, ?, ?, ?)"

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

		_, err := stmt.Exec(it.ID, it.Title, it.URL, deleted, dead, it.DiscussLink, it.ID, now, it.Domain, it.Description, it.By, it.Textx, it.EncryptedURL, it.EncryptedDiscussLink)
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
