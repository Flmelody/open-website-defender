package cache

import (
	"encoding/binary"

	"github.com/coocood/freecache"
)

// LocalCache implements Cache using freecache (in-process memory).
type LocalCache struct {
	store *freecache.Cache
}

// NewLocalCache creates a new local cache with the given memory size in megabytes.
func NewLocalCache(sizeMB int) *LocalCache {
	if sizeMB <= 0 {
		sizeMB = 100
	}
	return &LocalCache{
		store: freecache.NewCache(sizeMB * 1024 * 1024),
	}
}

func (c *LocalCache) Get(key string) ([]byte, error) {
	val, err := c.store.Get([]byte(key))
	if err != nil {
		return nil, ErrNotFound
	}
	return val, nil
}

func (c *LocalCache) Set(key string, value []byte, ttlSeconds int) error {
	return c.store.Set([]byte(key), value, ttlSeconds)
}

func (c *LocalCache) Del(key string) error {
	c.store.Del([]byte(key))
	return nil
}

func (c *LocalCache) Incr(key string, ttlSeconds int) (int64, error) {
	k := []byte(key)

	val, err := c.store.Get(k)
	var count int64
	if err == nil && len(val) == 8 {
		count = int64(binary.BigEndian.Uint64(val))
	}

	count++
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(count))

	ttl := ttlSeconds
	if err == nil {
		// Key exists: read remaining TTL to preserve the window.
		remaining, _ := c.store.TTL(k)
		if remaining > 0 {
			ttl = int(remaining)
		}
	}
	if err := c.store.Set(k, buf, ttl); err != nil {
		return 0, err
	}

	return count, nil
}

func (c *LocalCache) Clear() error {
	c.store.Clear()
	return nil
}

func (c *LocalCache) Close() error {
	return nil // freecache has no close
}
