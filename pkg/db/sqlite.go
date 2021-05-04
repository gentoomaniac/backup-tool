package db

import (
	"database/sql"
	"fmt"
	"strconv"

	_ "github.com/mattn/go-sqlite3" // blaa
	"github.com/rs/zerolog/log"
)

func NewSQLLite(dbpath string) (*SQLLiteDB, error) {
	rawDB, err := sql.Open("sqlite3", dbpath)
	return &SQLLiteDB{rawDB: rawDB}, err
}

type SQLLiteDB struct {
	rawDB *sql.DB
}

func (db *SQLLiteDB) runStatement(sql string) (sql.Result, error) {
	statement, err := db.rawDB.Prepare(sql)
	if err != nil {
		return nil, err
	}
	result, err := statement.Exec()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *SQLLiteDB) Init() (err error) {
	_, err = db.runStatement("PRAGMA foreign_keys = ON")
	if err != nil {
		return
	}
	log.Debug().Msg("Enabling foreign keys")

	db.runStatement(
		"CREATE TABLE IF NOT EXISTS blocks (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT, " +
			"hash BLOB, " +
			"name BLOB, " +
			"size INTEGER, " +
			"secret BLOB, " +
			"iv BLOB" +
			")")
	log.Debug().Msg("Created block table")

	db.runStatement(
		"CREATE TABLE IF NOT EXISTS fsobjects (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT, " +
			"name string, " +
			"path TEXT, " +
			"filemode INTEGER, " +
			"uid INTEGER, " +
			"gid INTEGER, " +
			"target TEXT, " +
			"hash BLOB" +
			")")
	log.Debug().Msg("Created fsobjects table")

	db.runStatement(
		"CREATE TABLE IF NOT EXISTS fileblocks (" +
			"ordernumber INTEGER, " +
			"fsobjectid INTEGER, " +
			"blockid INTEGER, " +
			"FOREIGN KEY(fsobjectid) REFERENCES fsobjects(id) ," +
			"FOREIGN KEY(blockid) REFERENCES blocks(id)" +
			")")
	log.Debug().Msg("Created fsobject<>block table")

	db.runStatement(
		"CREATE TABLE IF NOT EXISTS backups (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT, " +
			"name TEXT, " +
			"description TEXT, " +
			"blocksize INTEGER, " +
			"created INTEGER, " +
			"expires INTEGER" +
			")")
	log.Debug().Msg("Created backups table")

	db.runStatement(
		"CREATE TABLE IF NOT EXISTS backupobjects (" +
			"backupid INTEGER, " +
			"fsobjectid INTEGER, " +
			"FOREIGN KEY(backupid) REFERENCES backups(id) ," +
			"FOREIGN KEY(fsobjectid) REFERENCES fsobjects(id)" +
			")")
	log.Debug().Msg("Created backups<>fsobjetcs table")

	return err
}

func (db *SQLLiteDB) AddBlockToIndex(block *BlockMeta) error {
	_, err := db.rawDB.Exec("INSERT INTO blocks (hash, name, size, secret, iv) VALUES(?, ?, ?, ?, ?)", block.Hash, block.Name, block.Size, block.Secret, block.IV)
	return err
}

func (db *SQLLiteDB) GetBlockMeta(hash []byte) (*BlockMeta, error) {
	rows, err := db.rawDB.Query(fmt.Sprintf("SELECT * FROM blocks WHERE hash=x'%x'", hash))
	if err != nil {
		return nil, err
	}
	var bm BlockMeta
	for rows.Next() {
		rows.Scan(bm.ID, bm.Hash, bm.Name, bm.Size, bm.Secret, bm.IV)
		//fmt.Printf("%d | %x | %x | %d | %x | %x\n", id, rhash, name, size, secret, iv)
		rows.Close()
		return &bm, nil
	}
	return nil, nil
}

func (db *SQLLiteDB) AddFileToIndex(file *FSObject) error {
	_, err := db.rawDB.Exec("INSERT INTO fsobjects (name, path, filemode, uid, gid, target, hash) VALUES(?, ?, ?, ?, ?, ?, ?)",
		file.Name, file.Path, file.FileMode, file.User, file.Group, "", file.Hash)
	return err
}

func (db *SQLLiteDB) GetFSObj(name string, path string) ([]*FSObject, error) {
	rows, err := db.rawDB.Query(fmt.Sprintf("SELECT * FROM fsobjects WHERE name='%s' AND path='%s'", name, path))
	if err != nil {
		return nil, err
	}

	var obj *FSObject
	var objects []*FSObject
	objects = make([]*FSObject, 0)

	for rows.Next() {
		obj = &FSObject{}
		err := rows.Scan(&obj.ID, &obj.Name, &obj.Path, &obj.FileMode, &obj.User, &obj.Group, &obj.Target, &obj.Hash)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)

	}
	rows.Close()
	return objects, nil
}

func (db *SQLLiteDB) AddBackupToIndex(backup *Backup) error {
	_, err := db.rawDB.Exec("INSERT INTO backups (name, description, blocksize, created, expires) VALUES(?, ?, ?, ?, ?)",
		backup.Name, backup.Description, backup.Blocksize, backup.Timestamp, backup.Expiration)
	return err
}

func (db *SQLLiteDB) GetBackupBackupById(ID int) (*Backup, error) {
	rows, err := db.rawDB.Query("SELECT * FROM backups WHERE ID=" + strconv.FormatInt(int64(ID), 10))
	if err != nil {
		return nil, err
	}

	backup := &Backup{}
	for rows.Next() {
		rows.Scan(&ID, &backup.Name, &backup.Description, &backup.Blocksize, &backup.Timestamp, &backup.Expiration)
	}

	return backup, nil
}
