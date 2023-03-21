package lru

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type testValue struct {
	value string
}

func (t *testValue) Len() int {
	return len(t.value)
}

func TestCache_Get(t *testing.T) {
	c := New(100, nil)
	Convey("key not exists", t, func() {
		value, ok := c.Get("key")
		So(value, ShouldBeNil)
		So(ok, ShouldBeFalse)
	})
	Convey("key exists", t, func() {
		key, value := "key", &testValue{"value"}
		c.Add(key, value)
		got, ok := c.Get(key)
		So(got, ShouldEqual, value)
		So(ok, ShouldBeTrue)
	})
}

func TestCache_RemoveOldest(t *testing.T) {
	Convey("functional test", t, func() {
		c := New(100, func(key string, value Value) {
			fmt.Printf("key:%s, value:%v\n", key, value)
		})
		key1, value1 := "key1", &testValue{"value1"}
		key2, value2 := "key2", &testValue{"value2"}
		Convey("add in order", func() {
			c.Add(key1, value1)
			c.Add(key2, value2)
			c.RemoveOldest()

			got, ok := c.Get(key1)
			So(got, ShouldBeNil)
			So(ok, ShouldBeFalse)
			got, ok = c.Get(key2)
			So(got, ShouldEqual, value2)
			So(ok, ShouldBeTrue)
		})
		Convey("repeat add", func() {
			c.Add(key1, value1)
			c.Add(key2, value2)
			c.Add(key1, value1)
			c.RemoveOldest()

			got, ok := c.Get(key1)
			So(got, ShouldEqual, value1)
			So(ok, ShouldBeTrue)
			got, ok = c.Get(key2)
			So(got, ShouldBeNil)
			So(ok, ShouldBeFalse)
		})
	})
}

func TestCache_Len(t *testing.T) {
	Convey("functional test", t, func() {
		c := New(10, nil)
		key1, key2, value := "key1", "key2", &testValue{"value"}
		Convey("initial status", func() {
			So(c.Len(), ShouldEqual, 0)
		})
		Convey("add a unique key", func() {
			c.Add(key1, value)
			So(c.Len(), ShouldEqual, 1)
			c.Add(key2, value)
			So(c.Len(), ShouldEqual, 2)
		})
		Convey("add a existing key", func() {
			c.Add(key1, value)
			c.Add(key1, value)
			So(c.Len(), ShouldEqual, 1)
		})
	})
}
