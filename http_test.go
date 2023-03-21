package geecache

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
)

const (
	addr       = "localhost:9876"
	groupName  = "scores"
	cacheBytes = 1 << 10
)

var db map[string]string
var peers *HttpPool

func init() {
	db = map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	NewGroup(groupName, cacheBytes, GetterFunc(func(key string) ([]byte, error) {
		log.Printf("local search of key: %s\n", key)
		if value, ok := db[key]; ok {
			return []byte(value), nil
		}
		return nil, fmt.Errorf("%s not exists", key)
	}))
	peers = NewHttpPool(addr)

	go func() {
		log.Printf("geecache is running at %s\n", addr)
		log.Fatal(http.ListenAndServe(addr, peers))
	}()
}

func TestHttpPool_ServerHttp(t *testing.T) {
	Convey("functional test", t, func() {
		baseUrl, _ := url.Parse(fmt.Sprintf("http://%s", addr))
		Convey("unexpected path", func() {
			unexpectedUrl, _ := baseUrl.Parse("test")
			unexpectedPath := unexpectedUrl.String()
			resp, err := http.Get(unexpectedPath)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("bad request", func() {
			badRequestUrl, _ := baseUrl.Parse("_geecache/test")
			badRequestPath := badRequestUrl.String()
			resp, err := http.Get(badRequestPath)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusBadRequest)
		})
		Convey("group not found", func() {
			badRequestUrl, _ := baseUrl.Parse("_geecache/test/test")
			badRequestPath := badRequestUrl.String()
			resp, err := http.Get(badRequestPath)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusNotFound)
		})
		Convey("key not exists", func() {
			keyNotExists := "test"
			keyNotExistsUrl, _ := baseUrl.Parse(fmt.Sprintf("%s%s/%s", defaultBasePath, groupName, keyNotExists))
			keyNotExistsPath := keyNotExistsUrl.String()
			resp, err := http.Get(keyNotExistsPath)
			So(err, ShouldBeNil)
			So(resp.StatusCode, ShouldEqual, http.StatusInternalServerError)
		})
		Convey("key exists", func() {
			keyExists := "Tom"
			keyExistsUrl, _ := baseUrl.Parse(fmt.Sprintf("%s%s/%s", defaultBasePath, groupName, keyExists))
			keyExistsPath := keyExistsUrl.String()
			resp, err := http.Get(keyExistsPath)
			So(err, ShouldBeNil)
			defer resp.Body.Close()
			content, _ := ioutil.ReadAll(resp.Body)
			So(string(content), ShouldEqual, db[keyExists])
		})
	})
}

func TestHttpPool_PickPeer(t *testing.T) {
	Convey("functional test", t, func() {
		Convey("no peer exists", func() {
			key := "test"
			peer, ok := peers.PickPeer(key)
			So(peer, ShouldBeNil)
			So(ok, ShouldBeFalse)
		})

		peersToAdd := []string{"test", "debug"}
		peers.Set(peersToAdd...)

		Convey("key not exists", func() {
			// should return the nearest one
			keyNotExists := "release"
			peer, ok := peers.PickPeer(keyNotExists)
			So(peer, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
		Convey("key exists", func() {
			keyExists := "debug"
			peer, ok := peers.PickPeer(keyExists)
			So(peer, ShouldNotBeNil)
			So(ok, ShouldBeTrue)
		})
	})
}
