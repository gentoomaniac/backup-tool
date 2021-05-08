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

	_, err = db.runStatement(
		"CREATE TABLE IF NOT EXISTS blocks (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT, " +
			"hash BLOB, " +
			"name BLOB, " +
			"size INTEGER, " +
			"secret BLOB, " +
			"iv BLOB" +
			")")
	if err != nil {
		return err
	}

	_, err = db.runStatement(
		"CREATE TABLE IF NOT EXISTS fsobjects (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT, " +
			"name string, " +
			"path TEXT, " +
			"isdir INTEGER, " +
			"filemode INTEGER, " +
			"uid INTEGER, " +
			"gid INTEGER, " +
			"target TEXT, " +
			"hash BLOB, " +
			"backupid INTEGER, " +
			"FOREIGN KEY(backupid) REFERENCES backups(id)" +
			")")
	if err != nil {
		return err
	}

	_, err = db.runStatement(
		"CREATE TABLE IF NOT EXISTS fileblocks (" +
			"ordernumber INTEGER, " +
			"fsobjectid INTEGER, " +
			"blockid INTEGER, " +
			"FOREIGN KEY(fsobjectid) REFERENCES fsobjects(id) ," +
			"FOREIGN KEY(blockid) REFERENCES blocks(id)" +
			")")
	if err != nil {
		return err
	}

	_, err = db.runStatement(
		"CREATE TABLE IF NOT EXISTS backups (" +
			"id INTEGER PRIMARY KEY AUTOINCREMENT, " +
			"name TEXT, " +
			"description TEXT, " +
			"blocksize INTEGER, " +
			"created INTEGER, " +
			"expires INTEGER," +
			"UNIQUE(name)" +
			")")
	if err != nil {
		return err
	}

	return err
}

func (db *SQLLiteDB) AddBlockToIndex(block *BlockMeta) (int64, error) {
	result, err := db.rawDB.Exec("INSERT INTO blocks (hash, name, size, secret, iv) VALUES(?, ?, ?, ?, ?)",
		block.Hash, block.Name, block.Size, block.Secret, block.IV)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
}

func (db *SQLLiteDB) GetBlockMeta(hash []byte) (*BlockMeta, error) {
	rows, err := db.rawDB.Query(fmt.Sprintf("SELECT * FROM blocks WHERE hash=x'%x'", hash))
	if err != nil {
		return nil, err
	}
	var bm BlockMeta
	for rows.Next() {
		rows.Scan(&bm.ID, &bm.Hash, &bm.Name, &bm.Size, &bm.Secret, &bm.IV)
		rows.Close()
		return &bm, nil
	}
	return nil, nil
}

func (db *SQLLiteDB) AddFileToIndex(file *FSObject) (int64, error) {
	var isdir = 0
	if file.IsDir {
		isdir = 1
	}
	result, err := db.rawDB.Exec("INSERT INTO fsobjects (name, path, isdir, filemode, uid, gid, target, hash, backupid) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)",
		file.Name, file.Path, isdir, file.FileMode, file.User, file.Group, "", file.Hash, file.BackupID)
	if err != nil {
		return -1, err
	}

	return result.LastInsertId()
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
		var isdir int
		err := rows.Scan(&obj.ID, &obj.Name, &obj.Path, &isdir, &obj.FileMode, &obj.User, &obj.Group, &obj.Target, &obj.Hash, &obj.BackupID)
		if err != nil {
			return nil, err
		}
		if isdir > 0 {
			obj.IsDir = true
		}
		objects = append(objects, obj)

	}
	rows.Close()
	return objects, nil
}

func (db *SQLLiteDB) AddBackupToIndex(backup *Backup) (int64, error) {
	result, err := db.rawDB.Exec("INSERT INTO backups (name, description, blocksize, created, expires) VALUES(?, ?, ?, ?, ?)",
		backup.Name, backup.Description, backup.Blocksize, backup.Timestamp, backup.Expiration)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()
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

func (db *SQLLiteDB) GetBackups() (backups []*Backup, err error) {
	rows, err := db.rawDB.Query("SELECT * FROM backups ORDER BY created DESC")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		backup := &Backup{}
		rows.Scan(&backup.ID, &backup.Name, &backup.Description, &backup.Blocksize, &backup.Timestamp, &backup.Expiration)
		backups = append(backups, backup)
	}

	return backups, nil
}

func (db *SQLLiteDB) GetFSObjForBackups(id int64) (objects []*FSObject, err error) {
	rows, err := db.rawDB.Query(fmt.Sprintf("SELECT * FROM fsobjects WHERE backupid=%d", id))
	if err != nil {
		log.Debug().Err(err).Msg("")
		return nil, err
	}

	for rows.Next() {
		var isdir int
		obj := &FSObject{}
		err = rows.Scan(&obj.ID, &obj.Name, &obj.Path, &isdir, &obj.FileMode, &obj.User, &obj.Group, &obj.Target, &obj.Hash, &obj.BackupID)
		if err != nil {
			return
		}

		if isdir > 0 {
			obj.IsDir = true
		}
		objects = append(objects, obj)
	}
	return
}
