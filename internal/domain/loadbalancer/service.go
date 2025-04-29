package loadbalancer

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"loadbalancer/internal/pkg/config"
)

type Logegr interface {
	Infof(msg string, v ...interface{})
	Errorf(msg string, v ...interface{})
	Debugf(msg string, v ...interface{})
}

type LoadBalancer struct {
	mxBackends         sync.RWMutex
	Backends           []url.URL
	mxAliveBackends    sync.RWMutex
	AliveBackends      []url.URL
	TimeHealthCheck    int
	EndpointHealtCheck string
	Log                Logegr
}

func NewLoadBalancer(cfgBackends []config.Backend, cfgLoadBalancer config.LoadBalancer, l Logegr) *LoadBalancer {
	backends := make([]url.URL, 0, 0)

	for _, v := range cfgBackends {
		urlstr := fmt.Sprintf("%s://%s:%s", v.Scheme, v.Host, v.Port)
		url, err := url.Parse(urlstr)
		if err != nil {
			l.Errorf("Error parsing url %s", urlstr)
			continue
		}
		backends = append(backends, *url)
	}
	return &LoadBalancer{
		Backends:           backends,
		AliveBackends:      make([]url.URL, 0, 0),
		TimeHealthCheck:    cfgLoadBalancer.TimeHealthCheck,
		EndpointHealtCheck: cfgLoadBalancer.EndpointHealtCheck,
		Log:                l,
	}

}

func (s *LoadBalancer) NextBackend() *url.URL {
	if len(s.AliveBackends) == 0 {
		return nil
	}
	s.mxAliveBackends.Lock()
	defer s.mxAliveBackends.Unlock()
	target := s.AliveBackends[0]
	s.AliveBackends = append(s.AliveBackends[1:], target)
	return &target
}

func (s *LoadBalancer) setAliveBackends(backends []url.URL) {
	s.mxAliveBackends.Lock()
	defer s.mxAliveBackends.Unlock()
	s.AliveBackends = backends
}

func (s *LoadBalancer) RunHealthCheck(ctx context.Context) {
	go func(ctx context.Context) {
		for {
			if ctx.Err() != nil {
				break
			}
			aliveBackends := make([]url.URL, 0)
			for _, backend := range s.Backends {
				res, err := http.Get(backend.String() + s.EndpointHealtCheck)
				if err != nil || res.StatusCode != http.StatusOK {
					s.Log.Errorf("Backend %s is not alive", backend.String())
					continue
				}
				aliveBackends = append(aliveBackends, backend)
			}
			s.setAliveBackends(aliveBackends)
			s.Log.Infof("Health check finished. Alive backends: %v", aliveBackends)
			time.Sleep(time.Duration(s.TimeHealthCheck) * time.Second)
		}
	}(ctx)
}
