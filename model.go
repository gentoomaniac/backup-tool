package model

import (
	"os"
)

type Backup {
	Identifier string
	Size int
	Blocksize int
	Timestamp int
	Objects []*FSObject
}

func (b *Backup) Size() (size int){
	for object := range b.Objects {
		size += object.Size()
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
	Blocks []*Block
}

func (f *FSObject) Size() (size int){
	for file := range f.Blocks {
		size += f.Size
	}
	return
}

type Block struct {
	Hash      []byte
	Password  []byte
	Name 	  []byte
	Size      int
}