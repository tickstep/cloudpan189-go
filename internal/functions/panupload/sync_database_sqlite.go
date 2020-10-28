package panupload

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type FolderSyncDb struct {
	db     *sql.DB
	bucket []byte
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
func (db *FolderSyncDb) Get(key string) []byte {
	var data []byte
	row := db.db.QueryRow("select data from ecloud where path=?", key)
	row.Scan(&data)
	return data
}

func (db *FolderSyncDb) Put(key string, value []byte) error {
	var t int
	var err error
	row := db.db.QueryRow("select 1 from ecloud where path=?", key)
	row.Scan(&t)
	if t > 0 {
		_, err = db.db.Exec("update ecloud set data=? where path=?", value, key)
	} else {
		_, err = db.db.Exec("insert into ecloud values(?,?)", key, value)
	}
	return err
}

func (db *FolderSyncDb) Close() error {
	return db.db.Close()
}
