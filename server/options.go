package server

import (
	"context"
	"crypto/tls"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"
)

type Option func(*config)

type config struct {
	stopMonitoring []StopMonitoringFunc
	fiberCfg fiber.Config
	grpcOptions []grpc.ServerOption
	registries []ServiceRegistry[any]
	router Router
	tlsConfig *tls.Config
}

func newConfig(opts []Option) config {
	cfg := config{}

	for _, opt := range opts {
		opt(&cfg)
	}

	return cfg
}

type StopMonitoringFunc func(ctx context.Context)

type RegisterFunc[T any] func(grpc.ServiceRegistrar, T)

type ServiceRegistry[T any] struct {
	Service   T
	Registrar RegisterFunc[T]
}

type Router interface {
	RegisterRoutes(app *fiber.App)
}

func WithFiberConfig(cfg fiber.Config) Option {
	return func(c *config) {
		c.fiberCfg = cfg
	}
}

func WithGrpcConfig(cfg ...grpc.ServerOption) Option {
	return func(c *config) {
		c.grpcOptions = cfg
	}
}

func WithStopMonitor(fn StopMonitoringFunc) Option {
	return func(c *config) {
		c.stopMonitoring = append(c.stopMonitoring, fn)
	}
}

func WithGrpcRegistry[T any](service T, registrar RegisterFunc[T]) Option {
	wrappedRegistrar := func (s grpc.ServiceRegistrar, _ any)  {
		registrar(s, service)
	}

	return func(c *config) {
		c.registries = append(c.registries, ServiceRegistry[any]{
			Service: service,
			Registrar: wrappedRegistrar,
		})
	}
}

func WithGrpcRegistries(registries ...Option) Option {
	return func(c *config) {
		for _, opt := range registries {
			opt(c)
		}
	}
}

func WithRouter(r Router) Option {
	return func(c *config) {
		c.router = r
	}
}

func WithTLS(cfg *tls.Config) Option {
	return func(c *config) {
		c.tlsConfig = cfg
	}
}