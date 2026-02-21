package cache

import "errors"

// ErrNotFound is returned when a cache key does not exist.
var ErrNotFound = errors.New("cache: key not found")

// Cache defines the interface for cache backends.
// Implementations must be safe for concurrent use.
type Cache interface {
	// Get retrieves a value by key. Returns ErrNotFound if the key does not exist.
	Get(key string) ([]byte, error)

	// Set stores a value with a TTL in seconds. A TTL of 0 means no expiration.
	Set(key string, value []byte, ttlSeconds int) error

	// Del removes a key from the cache. No error if key does not exist.
	Del(key string) error

	// Incr atomically increments a counter by 1 and returns the new value.
	// If the key does not exist, it is initialized to 1 with the given TTL.
	// If the key already exists, its TTL is NOT reset (preserving the sliding window).
	Incr(key string, ttlSeconds int) (int64, error)

	// Clear removes all entries from the cache.
	Clear() error

	// Close releases any resources held by the cache backend.
	Close() error
}
