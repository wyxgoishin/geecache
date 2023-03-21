package geecache

import (
	"errors"
	"fmt"
	pb "geecache/geecachepb"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/protobuf/proto"
	"log"
	"testing"
)

var (
	keyThatMatchesNoPeer   = "debug"
	keyThatMatchesNoGetter = "test"
	keyThatCanBeGetByPeers = "release"
)

type presudoPeerPicker struct{}

func (p *presudoPeerPicker) PickPeer(key string) (PeerGetter, bool) {
	if key == keyThatMatchesNoPeer {
		return nil, false
	}
	return &presudoPeerGetter{}, true
}

type presudoPeerGetter struct{}

func (p *presudoPeerGetter) Get(req *pb.Request, resp *pb.Response) error {
	if req.GetKey() == keyThatMatchesNoGetter {
		return errors.New("unexpected group")
	}
	err := proto.Unmarshal([]byte(db[req.GetKey()]), resp)
	return err
}

func TestGetterFunc_Get(t *testing.T) {
	Convey("functional test", t, func() {
		var f Getter = GetterFunc(func(key string) ([]byte, error) {
			return []byte(key), nil
		})
		key := "key"
		got, err := f.Get(key)
		So(err, ShouldBeNil)
		So([]byte(key), ShouldResemble, got)
	})
}

func TestGet(t *testing.T) {
	Convey("benchmark test", t, func() {
		loadCounts := make(map[string]int, len(db))
		gee := NewGroup("scores", 2<<10, GetterFunc(
			func(key string) ([]byte, error) {
				log.Println("Get key:", key, "locally")
				if v, ok := db[key]; ok {
					loadCounts[key] += 1
					return []byte(v), nil
				}
				return nil, fmt.Errorf("%s not exist", key)
			}))

		Convey("get from local getter and local cache", func() {
			for key, value := range db {
				log.Println("Try to get key:", key)
				view, err := gee.Get(key)
				So(err, ShouldBeNil)
				So(view.String(), ShouldEqual, value)
				log.Println("Try to get key:", key)
				_, err = gee.Get(key)
				So(err, ShouldBeNil)
				So(loadCounts[key], ShouldEqual, 1)
			}
		})

		Convey("get unknown key", func() {
			view, err := gee.Get("unknown")
			So(err, ShouldNotBeNil)
			So(view, ShouldBeNil)
		})
	})
	Convey("functional test", t, func() {
		loadCounts := make(map[string]int, len(db))
		gee := NewGroup("scores", 2<<10, GetterFunc(
			func(key string) ([]byte, error) {
				log.Println("Get key:", key, "locally")
				if v, ok := db[key]; ok {
					loadCounts[key] += 1
					return []byte(v), nil
				}
				return nil, fmt.Errorf("%s not exist", key)
			}))

		Convey("no peers", func() {
			gee.Get("test")
		})

		peerPickers := &presudoPeerPicker{}
		gee.RegisterPeers(peerPickers)

		Convey("pick peer failed", func() {
			gee.Get(keyThatMatchesNoPeer)
		})
		Convey("get key from peer failed", func() {
			gee.Get(keyThatMatchesNoGetter)
		})
		Convey("get key from peer succeeded", func() {
			value, err := gee.Get(keyThatCanBeGetByPeers)
			So(value, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})
	})
}
