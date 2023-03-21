package consistentHash

import (
	"fmt"
	"hash/crc32"
	"sort"
)

type Hash func(data []byte) uint32

type Map struct {
	hashFunc   Hash
	replicas   uint
	hashKeys   []int // sorted, define as int as sort.Ints() is provided
	hashMap    map[string]bool
	hashKeyMap map[int]string
}

func New(replicas uint, hashFunc Hash) *Map {
	m := &Map{
		hashFunc:   hashFunc,
		replicas:   replicas,
		hashKeys:   make([]int, 0),
		hashMap:    make(map[string]bool),
		hashKeyMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		if m.hashMap[key] {
			continue
		}

		m.hashMap[key] = true
		for i := uint(0); i < m.replicas; i++ {
			hashKey := int(m.hashFunc([]byte(fmt.Sprintf("%s_%d", key, i))))
			m.hashKeys = append(m.hashKeys, hashKey)
			m.hashKeyMap[hashKey] = key
		}
	}
	sort.Ints(m.hashKeys)
}

// Get search the nearest key that is bigger than given key
func (m *Map) Get(key string) string {
	if len(m.hashKeys) == 0 {
		return ""
	}

	hashKey := int(m.hashFunc([]byte(key)))
	idx := sort.Search(len(m.hashKeys), func(i int) bool {
		return m.hashKeys[i] >= hashKey
	})

	return m.hashKeyMap[m.hashKeys[idx%len(m.hashKeys)]]
}
