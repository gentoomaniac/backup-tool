package sqlite

import (
	"database/sql"
	"fmt"
	"strconv"

	model "github.com/gentoomaniac/backup-tool/pkg/model"
	_ "github.com/mattn/go-sqlite3" // blaa
	"github.com/rs/zerolog/log"
)

func RunStatement(db *sql.DB, sql string) (sql.Result, error) {
	statement, err := db.Prepare(sql)
	if err != nil {
		return nil, err
	}
	result, err := statement.Exec()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func InitDB(dbpath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}
	_, err = RunStatement(db, "PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}
	log.Debug().Msg("Enabling foreign keys")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS blocks ("+
			"id INTEGER PRIMARY KEY AUTOINCREMENT, "+
			"hash BLOB, "+
			"name BLOB, "+
			"size INTEGER, "+
			"secret BLOB, "+
			"iv BLOB"+
			")")
	log.Debug().Msg("Created block table")

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
	log.Debug().Msg("Created fsobjects table")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS fileblocks ("+
			"ordernumber INTEGER, "+
			"fsobjectid INTEGER, "+
			"blockid INTEGER, "+
			"FOREIGN KEY(fsobjectid) REFERENCES fsobjects(id) ,"+
			"FOREIGN KEY(blockid) REFERENCES blocks(id)"+
			")")
	log.Debug().Msg("Created fsobject<>block table")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS backups ("+
			"id INTEGER PRIMARY KEY AUTOINCREMENT, "+
			"name TEXT, "+
			"description TEXT, "+
			"blocksize INTEGER, "+
			"created INTEGER, "+
			"expires INTEGER"+
			")")
	log.Debug().Msg("Created backups table")

	RunStatement(db,
		"CREATE TABLE IF NOT EXISTS backupobjects ("+
			"backupid INTEGER, "+
			"fsobjectid INTEGER, "+
			"FOREIGN KEY(backupid) REFERENCES backups(id) ,"+
			"FOREIGN KEY(fsobjectid) REFERENCES fsobjects(id)"+
			")")
	log.Debug().Msg("Created backups<>fsobjetcs table")

	return db, err
}

func AddBlockToIndex(db *sql.DB, block *model.BlockMeta) error {
	_, err := db.Exec("INSERT INTO blocks (hash, name, size, secret, iv) VALUES(?, ?, ?, ?, ?)", block.Hash, block.Name, block.Size, block.Secret, block.IV)
	return err
}

func GetBlockMeta(db *sql.DB, hash []byte) (*model.BlockMeta, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM blocks WHERE hash=x'%x'", hash))
	if err != nil {
		return nil, err
	}
	var bm model.BlockMeta
	for rows.Next() {
		rows.Scan(bm.ID, bm.Hash, bm.Name, bm.Size, bm.Secret, bm.IV)
		//fmt.Printf("%d | %x | %x | %d | %x | %x\n", id, rhash, name, size, secret, iv)
		rows.Close()
		return &bm, nil
	}
	return nil, nil
}

func AddFileToIndex(db *sql.DB, file *model.FSObject) error {
	_, err := db.Exec("INSERT INTO fsobjects (name, path, filemode, uid, gid, target, hash) VALUES(?, ?, ?, ?, ?, ?, ?)",
		file.Name, file.Path, file.FileMode, file.User, file.Group, "", file.Hash)
	return err
}

func GetFSObj(db *sql.DB, name string, path string) ([]*model.FSObject, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM fsobjects WHERE name='%s' AND path='%s'", name, path))
	if err != nil {
		return nil, err
	}

	var obj *model.FSObject
	var objects []*model.FSObject
	objects = make([]*model.FSObject, 0)

	for rows.Next() {
		obj = &model.FSObject{}
		err := rows.Scan(&obj.ID, &obj.Name, &obj.Path, &obj.FileMode, &obj.User, &obj.Group, &obj.Target, &obj.Hash)
		if err != nil {
			return nil, err
		}
		objects = append(objects, obj)

	}
	rows.Close()
	return objects, nil
}

func AddBackupToIndex(db *sql.DB, backup *model.Backup) error {
	_, err := db.Exec("INSERT INTO backups (name, description, blocksize, created, expires) VALUES(?, ?, ?, ?, ?)",
		backup.Name, backup.Description, backup.Blocksize, backup.Timestamp, backup.Expiration)
	return err
}

func GetBackupBackupById(db *sql.DB, ID int) (*model.Backup, error) {
	rows, err := db.Query("SELECT * FROM backups WHERE ID=" + strconv.FormatInt(int64(ID), 10))
	if err != nil {
		return nil, err
	}

	backup := &model.Backup{}
	for rows.Next() {
		rows.Scan(&ID, &backup.Name, &backup.Description, &backup.Blocksize, &backup.Timestamp, &backup.Expiration)
	}

	return backup, nil
}
