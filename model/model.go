package model

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
	Name     string
	Path     string
	Filetype *os.FileMode
	User     int64
	Group    int64
	Target   string
	Hash     []byte
	Blocks   []string
}

type BlockMeta struct {
	Hash   []byte
	Secret []byte
	Name   []byte
	Size   int
}
