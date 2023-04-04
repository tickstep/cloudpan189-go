package config

import (
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/library-go/expires"
	"path"
	"strconv"
	"time"
)

// DeleteCache 删除含有 dirs 的缓存
func (pu *PanUser) DeleteCache(dirs []string) {
	cache := pu.cacheOpMap.LazyInitCachePoolOp(strconv.FormatInt(pu.ActiveFamilyId, 10))
	for _, v := range dirs {
		key := v + "_" + "OrderByName"
		_, ok := cache.Load(key)
		if ok {
			cache.Delete(key)
		}
	}
}

// DeleteOneCache 删除缓存
func (pu *PanUser) DeleteOneCache(dirPath string) {
	ps := []string{dirPath}
	pu.DeleteCache(ps)
}

// CacheFilesDirectoriesList 缓存获取
func (pu *PanUser) CacheFilesDirectoriesList(pathStr string) (fdl *cloudpan.AppFileList, apiError *apierror.ApiError) {
	data := pu.cacheOpMap.CacheOperation(strconv.FormatInt(pu.ActiveFamilyId, 10), pathStr+"_OrderByName", func() expires.DataExpires {
		var fi *cloudpan.AppFileEntity
		fi, apiError = pu.panClient.AppFileInfoByPath(pu.ActiveFamilyId, pathStr)
		if apiError != nil {
			return nil
		}
		fileListParam := cloudpan.NewAppFileListParam()
		fileListParam.FileId = fi.FileId
		fileListParam.FamilyId = pu.ActiveFamilyId
		r, apiError := pu.panClient.AppGetAllFileList(fileListParam)
		if apiError != nil {
			return nil
		}
		// construct full path
		for _, f := range r.FileList {
			f.Path = path.Join(pathStr, f.FileName)
		}
		return expires.NewDataExpires(&r.FileList, 10*time.Minute)
	})
	if apiError != nil {
		return
	}
	return data.Data().(*cloudpan.AppFileList), nil
}
