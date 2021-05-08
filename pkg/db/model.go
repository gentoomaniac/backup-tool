package db

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Backup struct {
	ID          int64 `json:"id"`
	Blocksize   int   `json:"blocksize"`
	Timestamp   int   `json:"timestamp"`
	Objects     []*FSObject
	Name        string `json:"name`
	Description string `json:"description"`
	Expiration  int    `json:"expiration"`
}

func (b Backup) String() string {
	data, _ := json.Marshal(b)
	return string(data)
}

type FSObject struct {
	ID       int64
	Name     string
	Path     string
	IsDir    bool
	FileMode os.FileMode
	User     int
	Group    int
	Target   string
	Hash     []byte
	Blocks   []*BlockMeta
	BackupID int64
}

func (f FSObject) String() string {
	return fmt.Sprintf("%s %s %d:%d", path.Join(f.Path, f.Name), f.FileMode, f.User, f.Group)
}

type BlockMeta struct {
	ID     int64
	Hash   []byte
	Secret []byte
	IV     []byte
	Name   []byte
	Size   int
}
