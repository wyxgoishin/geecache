package geecache

import (
	pb "geecache/geecachepb"
	"geecache/singleflight"
	"log"
	"sync"
)

// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

// A Group is a cache namespace and associated data loaded spread over
type Group struct {
	name      string
	getter    Getter
	mainCache *cache
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	rw     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes uint64, getter Getter) *Group {
	if getter == nil {
		log.Fatal("nil getter")
	}

	rw.Lock()
	defer rw.Unlock()
	group := &Group{
		name:      name,
		getter:    getter,
		mainCache: &cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = group
	return group
}

func GetGroup(name string) *Group {
	rw.RLock()
	defer rw.RUnlock()
	return groups[name]
}

func (g *Group) Get(key string) (*ByteView, error) {
	if value, ok := g.mainCache.Get(key); ok {
		return value, nil
	}
	return g.load(key)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		log.Println("RegisterPeers called more than once")
		return
	}
	g.peers = peers
}

func (g *Group) load(key string) (*ByteView, error) {
	value, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				value, err := g.getFromPeer(peer, key)
				if err == nil {
					return value, nil
				}
				log.Printf("[GeeCache] failed to get from peer: %s", err.Error())
			}
		}
		return g.getLocally(key)
	})

	if err != nil {
		return nil, err
	}
	return value.(*ByteView), nil
}

func (g *Group) getLocally(key string) (*ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return nil, err
	}

	value := &ByteView{content: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (*ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	resp := &pb.Response{}
	if err := peer.Get(req, resp); err != nil {
		return nil, err
	}
	return &ByteView{content: cloneBytes(resp.GetValues())}, nil
}

func (g *Group) populateCache(key string, value *ByteView) {
	g.mainCache.Add(key, value)
}
