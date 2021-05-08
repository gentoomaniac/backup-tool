package db

import (
	"os"
)

type Backup struct {
	ID          int64
	Blocksize   int
	Timestamp   int
	Objects     []*FSObject
	Name        string
	Description string
	Expiration  int
}

type FSObject struct {
	ID       int64
	Name     string
	Path     string
	FileMode os.FileMode
	User     int
	Group    int
	Target   string
	Hash     []byte
	Blocks   []*BlockMeta
}

type BlockMeta struct {
	ID     int64
	Hash   []byte
	Secret []byte
	IV     []byte
	Name   []byte
	Size   int
}
