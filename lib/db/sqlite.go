package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/gentoomaniac/backup-tool/lib/model"
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
			"filemode INTEGER, "+
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

func AddBlockToIndex(db *sql.DB, block *model.BlockMeta) {
	_, err := db.Exec("INSERT INTO blocks (hash, name, size, secret, iv) VALUES(?, ?, ?, ?, ?)", block.Hash, block.Name, block.Size, block.Secret, block.IV)
	if err != nil {
		log.Error(err)
	}
	log.Debugf("Added block to index: %x", block.Hash)
}

func GetBlockMeta(db *sql.DB, hash []byte) *model.BlockMeta {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM blocks WHERE hash=x'%x'", hash))
	if err != nil {
		log.Error(err)
	}
	var bm model.BlockMeta
	for rows.Next() {
		rows.Scan(bm.ID, bm.Hash, bm.Name, bm.Size, bm.Secret, bm.IV)
		//fmt.Printf("%d | %x | %x | %d | %x | %x\n", id, rhash, name, size, secret, iv)
		rows.Close()
		return &bm
	}
	return nil
}

func AddFileToIndex(db *sql.DB, file *model.FSObject) {
	_, err := db.Exec("INSERT INTO fsobjects (name, path, filemode, uid, gid, target, hash) VALUES(?, ?, ?, ?, ?, ?, ?)",
		file.Name, file.Path, file.FileMode, file.User, file.Group, "", file.Hash)
	if err != nil {
		log.Error(err)
	}
	log.Debugf("Added file to index: %x", file.Hash)
}

func GetFSObj(db *sql.DB, name string, path string) []*model.FSObject {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM fsobjects WHERE name='%s' AND path='%s'", name, path))
	if err != nil {
		log.Error(err)
		return nil
	}

	var obj *model.FSObject
	var objects []*model.FSObject
	objects = make([]*model.FSObject, 0)

	for rows.Next() {
		obj = &model.FSObject{}
		err := rows.Scan(&obj.ID, &obj.Name, &obj.Path, &obj.FileMode, &obj.User, &obj.Group, &obj.Target, &obj.Hash)
		if err != nil {
			log.Error(err)
		} else {
			objects = append(objects, obj)
		}
	}
	rows.Close()
	return objects
}

func AddBackupToIndex(db *sql.DB, backup *model.Backup) {
	_, err := db.Exec("INSERT INTO backups (name, description, blocksize, created, expires) VALUES(?, ?, ?, ?, ?)",
		backup.Name, backup.Description, backup.Blocksize, backup.Timestamp, backup.Expiration)
	if err != nil {
		log.Error(err)
	}
	log.Debugf("Added backup to index: '%s'", backup.Name)
}
