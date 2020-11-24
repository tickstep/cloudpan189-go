// Copyright (c) 2020 tickstep & chenall
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package panupload

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tickstep/library-go/logger"

	jsoniter "github.com/json-iterator/go"

	_ "github.com/mattn/go-sqlite3"
)

type sqlite struct {
	*sql.DB
	bucket    []byte
	next      map[string]int
	cleanInfo *autoCleanInfo
}

func openSQLiteDB(file string, bucket string) (SyncDb, error) {
	db, err := sql.Open("sqlite3", file+"_sqlite.db")

	if err != nil {
		return nil, err
	}

	sqliteDb := &sqlite{DB: db, bucket: []byte(bucket), next: make(map[string]int)}

	if !sqliteDb.isTableExists("ecloud") {
		_, err = db.Exec("CREATE TABLE ecloud (path varchar PRIMARY KEY,data JSON);CREATE UNIQUE INDEX pk_path ON ecloud(path asc)")
		if err != nil {
			return nil, err
		}
	}

	logger.Verboseln("open sqlite db ok")

	return sqliteDb, nil
}

func (db *sqlite) isTableExists(table string) bool {
	var c int
	row := db.QueryRow(`SELECT 1 FROM sqlite_master WHERE type="table" AND name=?`, table)
	row.Scan(&c)
	return c == 1
}

func (db *sqlite) Get(key string) (ufm *UploadedFileMeta) {
	var data []byte
	ufm = &UploadedFileMeta{Path: key}
	row := db.QueryRow("select data from ecloud where path=?", key)
	if err := row.Scan(&data); err == nil {
		err = jsoniter.Unmarshal(data, ufm)
	}
	return ufm
}

func (db *sqlite) Put(key string, value *UploadedFileMeta) error {
	if db.cleanInfo != nil {
		value.LastSyncTime = db.cleanInfo.SyncTime
	}
	data, err := jsoniter.Marshal(value)
	if err != nil {
		return err
	}
	_, err = db.Exec("replace into ecloud values(?,?)", key, data)
	return err
}

func (db *sqlite) Del(key string) error {
	_, err := db.Exec("DELETE FROM ecloud where path=?", key)
	return err
}

func (db *sqlite) DelWithPrefix(prefix string) error {
	_, err := db.Exec("DELETE FROM ecloud where path like ?", prefix+"%")
	return err
}

func (db *sqlite) clean() (count uint) {
	syncFlag := fmt.Sprintf(`%%"synctime":%d%%`, db.cleanInfo.SyncTime)
	res, err := db.Exec(`DELETE FROM ecloud where path like ? AND data not like ?`, db.cleanInfo.PreFix+"%", syncFlag)
	if err != nil {
		return 0
	}
	var rowAffect int64
	rowAffect, err = res.RowsAffected()
	return uint(rowAffect)
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
	row := db.QueryRow("select rowid,path,data from ecloud where path like ? and rowid>?", prefix+"%", db.next[prefix])
	if err = row.Scan(&rowID, &meta.Path, &data); err == nil {
		err = jsoniter.Unmarshal(data, meta)
		db.next[prefix] = rowID
		return meta, nil
	}
	return nil, err
}

func (db *sqlite) AutoClean(prefix string, cleanFlag bool) {
	if !cleanFlag {
		db.cleanInfo = nil
	} else if db.cleanInfo == nil {
		db.cleanInfo = &autoCleanInfo{
			PreFix:   prefix,
			SyncTime: time.Now().Unix(),
		}
	}
}

func (db *sqlite) Close() error {
	if db.cleanInfo != nil {
		db.clean()
	}
	//数据库压缩清理
	db.Exec("VACUUM")
	return db.DB.Close()
}
