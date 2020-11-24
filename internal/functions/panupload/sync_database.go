package panupload

type SyncDb interface {
	//读取记录,返回值不会是nil
	Get(key string) (ufm *UploadedFileMeta)
	//删除单条记录
	Del(key string) error
	//根据前辍删除数据库记录，比如删除一个目录时可以连同子目录一起删除
	DelWithPrefix(prefix string) error
	Put(key string, value *UploadedFileMeta) error
	Close() error
	//读取数据库指定路径前辍的第一条记录（也作为循环获取的初始化，配置Next函数使用)
	First(prefix string) (*UploadedFileMeta, error)
	//获取指定路径前辍的的下一条记录
	Next(prefix string) (*UploadedFileMeta, error)
}

type SyncDBType uint8

const (
	DB_SQLITE SyncDBType = 1
	DB_NUTSDB SyncDBType = 2
)

func OpenSyncDb(t int, file string, bucket string) (SyncDb, error) {
	switch SyncDBType(t) {
	case DB_SQLITE:
	case DB_NUTSDB:
		return openNutsDb(file, bucket)
	}
	return openSQLiteDB(file, bucket)
}

type dbTableField struct {
	Path string
	Data []byte
}
