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
	if len(s.backends) <= 0 {
		return nil
	}
	newCurrent := s.current.Add(1)
	index := newCurrent % uint64(len(s.backends))
	return s.backends[index]
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	peer := s.GetNextPeer()
	if peer == nil {
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	peer.revProxy.ServeHTTP(w,r)
}
