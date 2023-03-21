package geecache

import (
	"fmt"
	"geecache/consistentHash"
	pb "geecache/geecachepb"
	"google.golang.org/protobuf/proto"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	defaultBasePath = "/_geecache/"
	defaultReplicas = 50
)

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(req *pb.Request, resp *pb.Response) error {
	group, key := req.GetGroup(), req.GetKey()
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	respHttp, err := http.Get(u)
	if err != nil {
		return err
	}

	if respHttp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %s", respHttp.Status)
	}

	bytes, err := ioutil.ReadAll(respHttp.Body)
	defer respHttp.Body.Close()
	if err != nil {
		return fmt.Errorf("reading response body: %s", err.Error())
	}

	if err = proto.Unmarshal(bytes, resp); err != nil {
		return fmt.Errorf("decoding response body: %v", err)
	}
	return nil
}

type HttpPool struct {
	selfPath string
	basePath string
	sync.Mutex
	peers       *consistentHash.Map
	httpGetters map[string]*httpGetter
}

func NewHttpPool(selfPath string) *HttpPool {
	return &HttpPool{
		selfPath: selfPath,
		basePath: defaultBasePath,
		peers:    consistentHash.New(defaultReplicas, nil),
	}
}

func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.selfPath, fmt.Sprintf(format, v...))
}

func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		http.Error(w, fmt.Sprintf("unexpected path: %s", r.URL.Path), http.StatusNotFound)
		return
	}

	p.Log("%s %s", r.Method, r.URL.Path)
	// <basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName, key := parts[0], parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, fmt.Sprintf("no such group: %s", groupName), http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func (p *HttpPool) Set(peers ...string) {
	p.Lock()
	defer p.Unlock()

	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}
	}
}

func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	p.Lock()
	defer p.Unlock()

	if peer := p.peers.Get(key); peer != "" && peer != p.selfPath {
		p.Log("pick peer: %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}
