package handlers

import (
	httperror "loadbalancer/internal/pkg/http"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type LoadBalancer interface {
	NextBackend() *url.URL
}

type Logger interface {
	Infof(msg string, v ...interface{})
	Errorf(msg string, v ...interface{})
	Debugf(msg string, v ...interface{})
}

type ReverseHandler struct {
	loadBalancer LoadBalancer
	log          Logger
}

func NewReverseHandler(lb LoadBalancer, logger Logger) *ReverseHandler {
	return &ReverseHandler{
		loadBalancer: lb,
		log:          logger,
	}
}

func (h *ReverseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := h.loadBalancer.NextBackend()
	if backend == nil {
		err := httperror.SendError(w, http.StatusServiceUnavailable, "no backend available")
		if err != nil {
			h.log.Errorf("failed to send error response: %v", err)
		}
		return
	}

	// Созадем новый прокси, устанавливая хост запроса. В случае ошибки возвращаем badgateway
	reverProxy := httputil.NewSingleHostReverseProxy(backend)
	reverProxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		h.log.Errorf("failed to proxy request: %v", err)
		err = httperror.SendError(w, http.StatusBadGateway, "failed to proxy request")
	}
	reverProxy.ServeHTTP(w, r)

}
