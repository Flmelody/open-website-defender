package cache

import (
	"open-website-defender/internal/infrastructure/logging"
	"sync"
)

var (
	store     Cache
	storeOnce sync.Once
)

// InitStore initializes the global cache store.
// Must be called once during application startup before any cache access.
func InitStore(sizeMB int) {
	storeOnce.Do(func() {
		if sizeMB <= 0 {
			sizeMB = 100
		}
		store = NewLocalCache(sizeMB)
		logging.Sugar.Infof("Cache initialized: %d MB", sizeMB)
	})
}

// Store returns the global cache instance.
func Store() Cache {
	if store == nil {
		// Fallback for tests or if InitStore was not called
		storeOnce.Do(func() {
			store = NewLocalCache(100)
		})
	}
	return store
}
