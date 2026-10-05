package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// A revision makes an in-flight read obsolete when a write invalidates its
// namespace. Include the DB identity to keep independent repositories isolated.
var readCacheState = struct {
	sync.Mutex
	revisions map[string]uint64
}{revisions: make(map[string]uint64)}

func cacheNamespace(db *sql.DB, name string) string {
	return fmt.Sprintf("read:%p:%s", db, name)
}

func invalidateReads(db *sql.DB, names ...string) {
	readCacheState.Lock()
	defer readCacheState.Unlock()
	for _, name := range names {
		readCacheState.revisions[cacheNamespace(db, name)]++
	}
}

// JSON snapshots give each caller its own pointers and slices. Errors are never
// cached. Old revision entries expire naturally, without a global cache flush.
func cachedRead[T any](db *sql.DB, namespace, key string, load func() (T, error)) (T, error) {
	if GlobalCache == nil {
		return load()
	}
	ns := cacheNamespace(db, namespace)
	readCacheState.Lock()
	revision := readCacheState.revisions[ns]
	readCacheState.Unlock()
	cacheKey := fmt.Sprintf("%s:%d:%s", ns, revision, key)
	result, err, _ := GlobalSFGroup.Do(cacheKey, func() (any, error) {
		if data, ok := GlobalCache.Get(cacheKey); ok {
			return data, nil
		}
		value, err := load()
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		readCacheState.Lock()
		if readCacheState.revisions[ns] == revision {
			GlobalCache.Set(cacheKey, data, 5*time.Second)
		}
		readCacheState.Unlock()
		return data, nil
	})
	var value T
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(result.([]byte), &value)
	return value, err
}
