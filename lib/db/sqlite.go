package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // blaa
	log "github.com/sirupsen/logrus"
)

func RunStatement(db *sql.DB, sql string) sql.Result {
	statement, err := db.Prepare(sql)
	if err != nil {
		log.Error(err)
	}
	result, err := statement.Exec()
	if err != nil {
		log.Error(err)
	}
	return result
}

func InitDB(dbpath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbpath)
	RunStatement(db, "PRAGMA foreign_keys = ON")
	log.Debug("Enabling foreign keys")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS blocks ("+
			"id INTEGER PRIMARY KEY AUTOINCREMENT, "+
			"hash BLOB, "+
			"name BLOB, "+
			"size INTEGER, "+
			"secret BLOB, "+
			"iv BLOB"+
			")")
	log.Debug("Created block table")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS fsobjects ("+
			"id INTEGER PRIMARY KEY AUTOINCREMENT, "+
			"name string, "+
			"path TEXT, "+
			"permissions INTEGER, "+
			"uid INTEGER, "+
			"gid INTEGER, "+
			"target TEXT, "+
			"hash BLOB"+
			")")
	log.Debug("Created fsobjects table")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS fileblocks ("+
			"ordernumber INTEGER, "+
			"fsobjectid INTEGER, "+
			"blockid INTEGER, "+
			"FOREIGN KEY(fsobjectid) REFERENCES fsobjects(id) ,"+
			"FOREIGN KEY(blockid) REFERENCES blocks(id)"+
			")")
	log.Debug("Created fsobject<>block table")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS backups ("+
			"id INTEGER PRIMARY KEY AUTOINCREMENT, "+
			"name TEXT, "+
			"description TEXT, "+
			"blocksize INTEGER, "+
			"created INTEGER, "+
			"expires INTEGER"+
			")")
	log.Debug("Created backups table")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS backupobjects ("+
			"backupid INTEGER, "+
			"fsobjectid INTEGER, "+
			"FOREIGN KEY(backupid) REFERENCES backups(id) ,"+
			"FOREIGN KEY(fsobjectid) REFERENCES fsobjects(id)"+
			")")
	log.Debug("Created backups<>fsobjetcs table")

	return db, err
}

func AddBlockToIndex(db *sql.DB, hash []byte, name []byte, size int, secret []byte, iv []byte) {
	_, err := db.Exec("INSERT INTO blocks (hash, name, size, secret, iv) VALUES(?, ?, ?, ?, ?)", hash, name, size, secret, iv)
	if err != nil {
		log.Error(err)
	}
	log.Debugf("Added block to index: %x", hash)
}

func IsBlockInIndex(db *sql.DB, hash []byte) bool {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM blocks WHERE hash=x'%x'", hash))
	if err != nil {
		log.Error(err)
	}
	var id int
	var rhash []byte
	var name []byte
	var size int
	var secret []byte
	var iv []byte
	for rows.Next() {
		rows.Scan(&id, &rhash, &name, &size, &secret, &iv)
		//fmt.Printf("%d | %x | %x | %d | %x | %x\n", id, rhash, name, size, secret, iv)
		rows.Close()
		return true
	}
	return false
}
