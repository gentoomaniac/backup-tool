package model

import (
	"os"
)

type Backup struct {
	Identifier string
	Blocksize  int
	Timestamp  int
	Objects    []*FSObject
}

func (b *Backup) GetSize() (size int) {
	for _, fsobj := range b.Objects {
		size += fsobj.Size()
	}
	return
}

type FSObject struct {
	Name     string
	Path     string
	Filetype *os.FileMode
	User     int64
	Group    int64
	Target   string
	Hash     []byte
	Blocks   []*BlockMeta
}

func (f *FSObject) Size() (size int) {
	for _, block := range f.Blocks {
		size += block.Size
	}
	return
}

type BlockMeta struct {
	Hash   []byte
	Secret []byte
	Name   []byte
	Size   int
}
