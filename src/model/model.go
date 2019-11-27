package model

import (
	"os"
)

type FSObject struct {
	Name     string
	Path     string
	Filetype *os.FileMode
	User     int64
	Group    int64
	Target   string
	Hash     []byte
	BlockIDs []*Block
}

type Block struct {
	Hash []byte
	Data []byte
}
