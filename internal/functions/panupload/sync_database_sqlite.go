//+build !nutsdb

package panupload

import (
	"database/sql"
	"fmt"

	jsoniter "github.com/json-iterator/go"

	_ "github.com/mattn/go-sqlite3"
)

type FolderSyncDb struct {
	db     *sql.DB
	bucket []byte
}

func init() {
	fmt.Println("sqlite")
}

func OpenSyncDb(file string, bucket string) (*FolderSyncDb, error) {
	db, err := sql.Open("sqlite3", file+"_sqlite.db")
	if err != nil {
		return nil, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS ecloud (path varchar PRIMARY KEY,data JSON)")
	if err != nil {
		return nil, err
	}

	return &FolderSyncDb{db: db, bucket: []byte(bucket)}, nil
}
func (db *FolderSyncDb) Get(key string) *UploadedFileMeta {
	var data []byte
	meta := &UploadedFileMeta{}
	row := db.db.QueryRow("select data from ecloud where path=?", key)
	if err := row.Scan(&data); err == nil {
		err = jsoniter.Unmarshal(data, meta)
	}
	return meta
}

func (db *FolderSyncDb) Put(key string, value *UploadedFileMeta) error {
	data, err := jsoniter.Marshal(value)
	if err != nil {
		return err
	}
	_, err = db.db.Exec("replace into ecloud values(?,?)", key, data)
	return err
}

func (db *FolderSyncDb) Close() error {
	return db.db.Close()
}
