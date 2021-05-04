package db

type DB interface {
	Init() error
	AddBlockToIndex(block *BlockMeta) error
	GetBlockMeta(hash []byte) (*BlockMeta, error)
	AddFileToIndex(file *FSObject) error
	GetFSObj(name string, path string) ([]*FSObject, error)
	AddBackupToIndex(backup *Backup) error
	GetBackupBackupById(ID int) (*Backup, error)
}
