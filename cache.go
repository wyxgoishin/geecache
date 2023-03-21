package geecache

import (
	"geecache/lru"
	"sync"
)

type cache struct {
	sync.Mutex
	lru        *lru.Cache
	cacheBytes uint64
}

func (c *cache) Add(key string, value *ByteView) {
	c.Lock()
	defer c.Unlock()

	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) Get(key string) (value *ByteView, ok bool) {
	// c.lru.Get will change lru, thus here should use Mutex rather than RWMutex
	c.Lock()
	defer c.Unlock()

	if c.lru == nil {
		return
	}

	if val, ret := c.lru.Get(key); ret {
		return val.(*ByteView), ret
	} else {
		return
	}
}
