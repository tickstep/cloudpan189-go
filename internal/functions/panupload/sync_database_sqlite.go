//+build !nutsdb

package panupload

import (
	"database/sql"

	"github.com/tickstep/library-go/logger"

	jsoniter "github.com/json-iterator/go"

	_ "github.com/mattn/go-sqlite3"
)

type sqlite struct {
	db     *sql.DB
	bucket []byte
	next   map[string]int
}

func isTableExists(db *sql.DB, table string) bool {
	var c int
	row := db.QueryRow(`SELECT 1 FROM sqlite_master WHERE type="table" AND name=?`, table)
	row.Scan(&c)
	return c == 1
}

func openSQLiteDB(file string, bucket string) (SyncDb, error) {
	db, err := sql.Open("sqlite3", file+"_sqlite.db")
	if err != nil {
		return nil, err
	}

	if !isTableExists(db, "ecloud") {
		_, err = db.Exec("CREATE TABLE ecloud (path varchar PRIMARY KEY,data JSON);CREATE UNIQUE INDEX pk_path ON ecloud(path asc)")
		if err != nil {
			return nil, err
		}
	}
	logger.Verboseln("open sqlite db ok")
	return &sqlite{db: db, bucket: []byte(bucket), next: make(map[string]int)}, nil
}
func (db *sqlite) Get(key string) (ufm *UploadedFileMeta) {
	var data []byte
	ufm = &UploadedFileMeta{Path: key}
	row := db.db.QueryRow("select data from ecloud where path=?", key)
	if err := row.Scan(&data); err == nil {
		err = jsoniter.Unmarshal(data, ufm)
	}
	return ufm
}

func (db *sqlite) Put(key string, value *UploadedFileMeta) error {
	data, err := jsoniter.Marshal(value)
	if err != nil {
		return err
	}
	_, err = db.db.Exec("replace into ecloud values(?,?)", key, data)
	return err
}

func (db *sqlite) Del(key string) error {
	_, err := db.db.Exec("DELETE FROM ecloud where path=?", key)
	return err
}

func (db *sqlite) DelWithPrefix(prefix string) error {
	_, err := db.db.Exec("DELETE FROM ecloud where path like ?", prefix+"%")
	return err
}

//读取数据库的第一条记录（也作为循环获取的初始化，配置Next函数使用)
func (db *sqlite) First(prefix string) (*UploadedFileMeta, error) {
	db.next[prefix] = -1
	return db.Next(prefix)
}

//获取数据库的下一条记录.
func (db *sqlite) Next(prefix string) (*UploadedFileMeta, error) {
	var data []byte
	var err error
	var rowID int
	meta := &UploadedFileMeta{}
	row := db.db.QueryRow("select rowid,path,data from ecloud where path like ? and rowid>?", prefix+"%", db.next[prefix])
	if err = row.Scan(&rowID, &meta.Path, &data); err == nil {
		err = jsoniter.Unmarshal(data, meta)
		db.next[prefix] = rowID
		return meta, nil
	}
	return nil, err
}

func (db *sqlite) Close() error {
	return db.db.Close()
}
