package middlewares

import (
	httperror "loadbalancer/internal/pkg/http"
	"net/http"
	"strings"
)

type Logger interface {
	Infof(msg string, v ...interface{})
	Errorf(msg string, v ...interface{})
	Debugf(msg string, v ...interface{})
}

type RateLimiter interface {
	CheckLimiter(clientIP string) bool
}

type Limiter struct {
	h           http.Handler
	rateLimiter RateLimiter
	log         Logger
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewLimiterMiddlewares(h http.Handler, rateLimiter RateLimiter, log Logger) http.Handler {
	return &Limiter{
		h:           h,
		rateLimiter: rateLimiter,
		log:         log,
	}
}

func (m *Limiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ip := strings.Split(r.RemoteAddr, ":")[0]
	m.log.Infof("request from %s", ip)
	if !m.rateLimiter.CheckLimiter(ip) {
		m.log.Infof("rate limit exceeded for %s", ip)
		httperror.SendError(w, http.StatusTooManyRequests, "Rate limit exceeded")
		return
	}
	m.h.ServeHTTP(w, r)
}
