package proxy

import (
	"log"
	"net"
	"net/url"
	"time"
)

// defer delays the execution of a function until the sorrounding function is finished

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.alive = alive
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.alive
}

func isBackendAlive(u *url.URL) bool {
	conn, err := net.DialTimeout("tcp", u.Host, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func (s *ServerPool) HealthCheck() {
	for _, b := range s.backends {
		if isBackendAlive(b.url) {
			b.SetAlive(true)
			log.Printf("Backend %s is UP", b.url.Host)
		} else {
			b.SetAlive(false)
			log.Printf("Backend %s is DOWN", b.url.Host)
		}
	}
}

func (s *ServerPool) StartHealthCheck(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		s.HealthCheck()
	}
}
