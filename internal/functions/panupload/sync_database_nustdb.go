//+build nutsdb

package panupload

import (
	jsoniter "github.com/json-iterator/go"

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
func (db *FolderSyncDb) Get(key string) *UploadedFileMeta {
	data := &UploadedFileMeta{}
	db.db.View(func(tx *nutsdb.Tx) error {
		ent, err := tx.Get(db.bucket, []byte(key))
		if err != nil {
			return err
		}
		return jsoniter.Unmarshal(ent.Value, &data)
	})

	return data
}

func (db *FolderSyncDb) Put(key string, value *UploadedFileMeta) error {
	return db.db.Update(func(tx *nutsdb.Tx) error {
		data, err := jsoniter.Marshal(value)
		if err != nil {
			return err
		}
		return tx.Put(db.bucket, []byte(key), data, 0)
	})
}

func (db *FolderSyncDb) Close() error {
	return db.db.Close()
}
