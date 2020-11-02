//+build !sqlite

package panupload

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/xujiajun/nutsdb"
)

type FolderSyncDb struct {
	db     *nutsdb.DB
	bucket string
}

func OpenSyncDb(file string, bucket string) (*FolderSyncDb, error) {
	opt := nutsdb.DefaultOptions
	opt.Dir = file
	opt.EntryIdxMode = nutsdb.HintBPTSparseIdxMode
	db, err := nutsdb.Open(opt)
	if err != nil {
		return nil, err
	}
	return &FolderSyncDb{db: db, bucket: bucket}, nil
}
func (db *FolderSyncDb) Get(key string) []byte {
	var data []byte
	db.db.View(func(tx *nutsdb.Tx) error {
		ent, err := tx.Get(db.bucket, []byte(key))
		if err != nil {
			return err
		}
		data = ent.Value
		return nil
	})
	return data
}

func (db *FolderSyncDb) Put(key string, value []byte) error {
	return db.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Put(db.bucket, []byte(key), value, 0)
	})
}

func (db *FolderSyncDb) Close() error {
	return db.db.Close()
}
