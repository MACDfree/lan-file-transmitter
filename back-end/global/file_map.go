package global

import "lft/cache"

// FileMap 文件列表
var FileMap *cache.Cache

func init() {
	FileMap = cache.New(1024, nil)
}
