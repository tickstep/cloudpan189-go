package panupload

import (
	"github.com/tickstep/library-go/logger"

	jsoniter "github.com/json-iterator/go"

	_ "github.com/mattn/go-sqlite3"
	"github.com/xujiajun/nutsdb"
)

type nutsDB struct {
	db     *nutsdb.DB
	bucket string
	next   nutsDBScan
}

type nutsDBScan struct {
	entries nutsdb.Entries
	off     int
	size    int
}

func openNutsDb(file string, bucket string) (SyncDb, error) {
	opt := nutsdb.DefaultOptions
	opt.Dir = file
	opt.EntryIdxMode = nutsdb.HintBPTSparseIdxMode
	db, err := nutsdb.Open(opt)
	if err != nil {
		return nil, err
	}
	logger.Verboseln("open nutsDb ok")
	return &nutsDB{db: db, bucket: bucket, next: nutsDBScan{}}, nil
}

func (db *nutsDB) Get(key string) (data *UploadedFileMeta) {
	data = &UploadedFileMeta{Path: key}
	db.db.View(func(tx *nutsdb.Tx) error {
		ent, err := tx.Get(db.bucket, []byte(key))
		if err != nil {
			return err
		}
		return jsoniter.Unmarshal(ent.Value, data)
	})

	return data
}

func (db *nutsDB) Del(key string) error {
	return db.db.Update(func(tx *nutsdb.Tx) error {
		return tx.Delete(db.bucket, []byte(key))
	})
}

func (db *nutsDB) DelWithPrefix(prefix string) error {
	return db.db.Update(func(tx *nutsdb.Tx) error {
		offset := 0
		for {
			ent, _, err := tx.PrefixScan(db.bucket, []byte(prefix), offset, 1)
			if err != nil {
				break
			}
			for _, item := range ent {
				tx.Delete(db.bucket, item.Key)
			}
			offset += 1
		}
		return nil
	})
}

func (db *nutsDB) First(prefix string) (*UploadedFileMeta, error) {
	db.db.View(func(tx *nutsdb.Tx) error {
		entries, _, err := tx.PrefixScan(db.bucket, []byte(prefix), 0, 0xffffffff)
		if err != nil {
			return err
		}
		db.next.entries = entries
		db.next.off = 0
		db.next.size = len(entries)
		return nil
	})
	return db.Next(prefix)
}

func (db *nutsDB) Next(prefix string) (*UploadedFileMeta, error) {
	data := &UploadedFileMeta{}
	for { //循环读取直到找到符合条件的记录
		if db.next.off >= db.next.size {
			return nil, nutsdb.ErrPrefixScansNoResult
		}
		ent := db.next.entries[db.next.off]
		db.next.off++
		if len(ent.Value) > 0 {
			jsoniter.Unmarshal(ent.Value, &data)
			data.Path = string(ent.Key)
			return data, nil
		}
	}
}

func (db *nutsDB) Put(key string, value *UploadedFileMeta) error {
	return db.db.Update(func(tx *nutsdb.Tx) error {
		data, err := jsoniter.Marshal(value)
		if err != nil {
			return err
		}
		return tx.Put(db.bucket, []byte(key), data, 0)
	})
}

func (db *nutsDB) Close() error {
	return db.db.Close()
}
