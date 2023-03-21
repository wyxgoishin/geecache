package consistentHash

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const (
	replicas uint = 10
)

func TestMap_Add(t *testing.T) {
	m := New(replicas, nil)
	Convey("functional test", t, func() {
		key1, key2 := "key1", "key2"
		Convey("duplicate key provided", func() {
			m.Add(key1, key1)
			So(len(m.hashKeys), ShouldEqual, replicas)
		})
		Convey("different key provided", func() {
			m.Add(key1, key2)
			So(len(m.hashKeys), ShouldEqual, 2*replicas)
		})
	})
}

func TestMap_Get(t *testing.T) {
	m := New(replicas, nil)
	Convey("functional test", t, func() {
		key1, key2, key3 := "key1", "key4", "key2"
		m.Add(key1, key3)
		key := m.Get(key2)
		So(key, ShouldEqual, key1)
	})
}
