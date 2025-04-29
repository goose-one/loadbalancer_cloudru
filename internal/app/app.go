package app

import (
	"context"
	"fmt"
	"loadbalancer/internal/app/handlers"
	"loadbalancer/internal/domain/loadbalancer"
	"loadbalancer/internal/domain/ratelimiter/repository"
	ratelimiter "loadbalancer/internal/domain/ratelimiter/service"
	"loadbalancer/internal/pkg/config"
	"loadbalancer/internal/pkg/http/middlewares"
	"loadbalancer/internal/pkg/logger"
	"net"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	cfg    *config.Config
	server http.Server
	ctx    context.Context
	cancel context.CancelFunc
}

func NewApp(cfgPath string) (*App, error) {
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
	}
	logger := logger.NewZerologLogger(cfg.Logger)
	logger.Infof("%v", cfg.Backends)
	loadbalancer := loadbalancer.NewLoadBalancer(cfg.Backends, cfg.LoadBalancer, logger)
	loadbalancer.RunHealthCheck(ctx)
	pool, err := getPoolDB(ctx, &cfg.DB)
	if err != nil {
		return nil, err
	}
	repository := repository.NewSQLRepository(pool)
	ratelimiter := ratelimiter.NewRateLimiter(repository, cfg.RateLimiter.MaxRequests, cfg.RateLimiter.Interval, logger)
	handlers := setupHandlers(loadbalancer, ratelimiter, logger)
	app.server.Handler = handlers

	return app, nil
}

func getPoolDB(ctx context.Context, cfg *config.Database) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (a *App) Run() error {
	address := fmt.Sprintf(":%s", a.cfg.Server.Port)

	l, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return a.server.Serve(l)
}

func (a *App) Shutdown() error {
	err := a.server.Shutdown(a.ctx)
	if err != nil {
		return err
	}
	a.cancel()
	return nil
}

func setupHandlers(lb handlers.LoadBalancer, rl ratelimiter.RateLimiterI, logger handlers.Logger) http.Handler {
	sm := http.NewServeMux()

	sm.Handle("/", handlers.NewReverseHandler(lb, logger))
	sm.Handle("POST /clients", handlers.NewAddClientHandler(rl, logger))
	mw := middlewares.NewLimiterMiddlewares(sm, rl, logger)
	return mw
}
