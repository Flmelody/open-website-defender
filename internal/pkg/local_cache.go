package pkg

import (
	"github.com/coocood/freecache"
	"sync"
)

var (
	instance *freecache.Cache
	once     sync.Once
)

func Cacher() *freecache.Cache {
	once.Do(func() {
		instance = freecache.NewCache(100 * 1024 * 1024)
	})
	return instance
}
