package global

import (
	"sync"
	"sync/atomic"
	"log"
)

var (
	Height      int64
	MemKeyCache *MemCache
)

func init() {
	MemKeyCache = NewMemCache()
}

func StoreHeight(height int64) {
	atomic.StoreInt64(&Height, height)
}

func LoadHeight() int64 {
	return atomic.LoadInt64(&Height)
}

type (
	MemCache struct {
		lock  *sync.RWMutex
		Cache map[int64]map[string]interface{}
	}
)

func NewMemCache() *MemCache {
	return &MemCache{
		lock:  new(sync.RWMutex),
		Cache: make(map[int64]map[string]interface{}),
	}
}

func (m *MemCache) GetCache(height int64, hash string) interface{} {
	//m.lock.RLock()
	//defer m.lock.RUnlock()

	hc := m.Cache[height]
	v, ok := hc[hash]
	if !ok {
		log.Println("not get cache", height, hash, len(hash))
		return nil
	}
	return v
}

func (m *MemCache) SetCache(height int64, hash string, v interface{}) {
	//m.lock.Lock()
	//defer m.lock.Unlock()

	_, ok := m.Cache[height]
	if !ok {
		m.Cache[height] = make(map[string]interface{})
	}
	m.Cache[height][hash] = v
}
