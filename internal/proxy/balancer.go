package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	url      *url.URL
	alive    bool
	mux      sync.RWMutex
	revProxy *httputil.ReverseProxy
}

type ServerPool struct {
	backends []*Backend
	current  atomic.Uint64
}

func NewBackend(rawURL string) (*Backend, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	return &Backend{
		url:      parsedURL,
		alive:    true,
		revProxy: proxy,
	}, nil
}

func NewServerPool(backends []*Backend) *ServerPool {
	return &ServerPool{
		backends: backends,
	}
}

func (s *ServerPool) GetNextPeer() *Backend {
	n := uint64(len(s.backends))
	if n <= 0 {
		return nil
	}
	next := s.current.Add(1)
	for i := 0; i < int(n); i++ {
		idx := (next + uint64(i)) % n
		if s.backends[idx].IsAlive() {
			return s.backends[idx]
		}
	}
	return nil
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := s.GetNextPeer()
	if peer == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	peer.revProxy.ServeHTTP(w, r)
}
