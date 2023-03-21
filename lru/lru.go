package lru

import "container/list"

type Cache struct {
	maxBytes  uint64
	nBytes    uint64
	lst       *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type Entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

func New(maxBytes uint64, onEvicted func(key string, value Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		lst:       list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if elem, ok := c.cache[key]; ok {
		c.lst.MoveToFront(elem)
		kv := elem.Value.(*Entry)
		return kv.value, ok
	}
	return
}

func (c *Cache) RemoveOldest() {
	if elem := c.lst.Back(); elem != nil {
		c.lst.Remove(elem)
		kv := elem.Value.(*Entry)
		delete(c.cache, kv.key)
		c.nBytes -= uint64(len(kv.key) + kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if elem, ok := c.cache[key]; ok {
		kv := elem.Value.(*Entry)
		c.nBytes += uint64(value.Len() - kv.value.Len())
		c.lst.MoveToFront(elem)
		kv.value = value
	} else {
		kv := &Entry{
			key:   key,
			value: value,
		}
		c.lst.PushFront(kv)
		c.cache[key] = c.lst.Front()
	}
	for c.nBytes > c.maxBytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.lst.Len()
}
