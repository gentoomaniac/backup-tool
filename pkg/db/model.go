package db

import (
	"os"
)

type Backup struct {
	Blocksize   int
	Timestamp   int
	Objects     []*FSObject
	Name        string
	Description string
	Expiration  int
}

type FSObject struct {
	ID       int
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
	ID     int
	Hash   []byte
	Secret []byte
	IV     []byte
	Name   []byte
	Size   int
}
