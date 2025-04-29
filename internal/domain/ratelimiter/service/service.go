package ratelimiter

import (
	"context"
	"fmt"
	"loadbalancer/internal/domain/models"
	"sync"
	"time"
)

type Logger interface {
	Infof(msg string, v ...interface{})
	Errorf(msg string, v ...interface{})
	Debugf(msg string, v ...interface{})
}

type Repository interface {
	AddNewClient(ctx context.Context, client models.Client) (int64, error)
	GetClinet(ctx context.Context, ip string) (*models.Client, error)
}

type RateLimiterI interface {
	CheckLimiter(clientIP string) bool
	UpdateClientConfig(ctx context.Context, client models.Client) (int64, error)
}

type RateLimiter struct {
	rep                 Repository
	defaultLimitRequest int
	defaultTimeTicker   int
	mx                  sync.RWMutex
	clients             map[string]*ClientBucket
	log                 Logger
}

type ClientBucket struct {
	Bucket chan struct{}
	Ctx    context.Context
	Cancel context.CancelFunc
}

func NewRateLimiter(rep Repository, defaultLimitRequest int, defaultTimeTicker int, log Logger) *RateLimiter {
	return &RateLimiter{
		rep:                 rep,
		defaultLimitRequest: defaultLimitRequest,
		defaultTimeTicker:   defaultTimeTicker,
		clients:             make(map[string]*ClientBucket),
		log:                 log,
	}
}

func (s *RateLimiter) CheckLimiter(clientIP string) bool {
	s.mx.Lock()
	defer s.mx.Unlock()
	valid := false
	s.log.Debugf("Checking client %s", clientIP)
	if _, ok := s.clients[clientIP]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		cfgClient := s.GetConfigClient(ctx, clientIP)
		if cfgClient == nil {
			cfgClient = &models.Client{
				IP:         clientIP,
				Capacity:   s.defaultLimitRequest,
				RatePerSec: s.defaultTimeTicker,
			}
			go s.AddConfigClient(ctx, *cfgClient)

		}
		s.clients[cfgClient.IP] = &ClientBucket{
			Bucket: make(chan struct{}, cfgClient.Capacity),
			Ctx:    ctx,
			Cancel: cancel,
		}
		s.log.Infof("New client %s", cfgClient.IP)
		s.log.Infof("New client %v", cfgClient)
		for i := 0; i < cfgClient.Capacity; i++ {
			s.clients[cfgClient.IP].Bucket <- struct{}{}
		}
		go replenishBucket(ctx, s.clients[cfgClient.IP].Bucket, cfgClient.Capacity, time.Duration(cfgClient.RatePerSec))
	}

	s.log.Debugf("Client %s has %d requests left", clientIP, len(s.clients[clientIP].Bucket))
	if len(s.clients[clientIP].Bucket) > 0 {
		<-s.clients[clientIP].Bucket
		valid = true
	}
	return valid
}

func (s *RateLimiter) GetConfigClient(ctx context.Context, ip string) *models.Client {
	client, err := s.rep.GetClinet(ctx, ip)
	if err != nil {
		s.log.Errorf("Error getting client %s", ip)
		return nil
	}
	return client
}

func (s *RateLimiter) AddConfigClient(ctx context.Context, client models.Client) {
	id, err := s.rep.AddNewClient(ctx, client)
	if err != nil {
		s.log.Errorf("Error add new client client %s", client.IP)
	}
	s.log.Infof("Add new client confif with id %d", id)
}

func (s *RateLimiter) UpdateClientConfig(ctx context.Context, client models.Client) (int64, error) {
	id, err := s.rep.AddNewClient(ctx, client)
	if err != nil {
		s.log.Errorf("Error getting client %s error: %v", client.IP, err)
		return 0, err
	}
	s.changeConfig(client)
	return id, nil

}

func (s *RateLimiter) changeConfig(client models.Client) {
	ctx, cancel := context.WithCancel(context.Background())
	s.mx.Lock()
	defer s.mx.Unlock()
	v, ok := s.clients[client.IP]
	if ok {
		v.Cancel()
		delete(s.clients, client.IP)
	}
	s.clients[client.IP] = &ClientBucket{
		Bucket: make(chan struct{}, client.Capacity),
		Ctx:    ctx,
		Cancel: cancel,
	}
	s.log.Infof("New client %s", client.IP)
	for i := 0; i < client.Capacity; i++ {
		s.clients[client.IP].Bucket <- struct{}{}
	}
	go replenishBucket(ctx, s.clients[client.IP].Bucket, client.Capacity, time.Duration(client.RatePerSec))

}

func replenishBucket(ctx context.Context, bucket chan struct{}, capacity int, rateTimeTicker time.Duration) {
	ticker := time.NewTicker(rateTimeTicker * time.Second)
	for range ticker.C {
		if ctx.Err() != nil {
			fmt.Println("CLOSE CHANEL")
			close(bucket)
			ticker.Stop()
			return
		}
		lenb := len(bucket)
		if lenb < capacity {
			fmt.Printf("LEN BUCKET %d \n", lenb)
			fmt.Printf("CAPACITY BUCKET %d \n", capacity)

			for i := 0; i < (capacity - lenb); i++ {
				bucket <- struct{}{}
			}
		}
	}
}
