package options

import (
	"context"
	"crypto/tls"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"
)

type Option func(*Config)

type Config struct {
	StopMonitoring []StopMonitoringFunc
	FiberCfg fiber.Config
	GrpcOptions []grpc.ServerOption
	Registries []ServiceRegistry[any]
	Router Router
	TlsConfig *tls.Config
}

func NewConfig(opts []Option) Config {
	cfg := Config{}

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
	return func(c *Config) {
		c.FiberCfg = cfg
	}
}

func WithGrpcConfig(cfg ...grpc.ServerOption) Option {
	return func(c *Config) {
		c.GrpcOptions = cfg
	}
}

func WithStopMonitor(fn StopMonitoringFunc) Option {
	return func(c *Config) {
		c.StopMonitoring = append(c.StopMonitoring, fn)
	}
}

func WithGrpcRegistry[T any](service T, registrar RegisterFunc[T]) Option {
	wrappedRegistrar := func (s grpc.ServiceRegistrar, _ any)  {
		registrar(s, service)
	}

	return func(c *Config) {
		c.Registries = append(c.Registries, ServiceRegistry[any]{
			Service: service,
			Registrar: wrappedRegistrar,
		})
	}
}

func WithGrpcRegistries(registries ...Option) Option {
	return func(c *Config) {
		for _, opt := range registries {
			opt(c)
		}
	}
}

func WithRouter(r Router) Option {
	return func(c *Config) {
		c.Router = r
	}
}

func WithTLS(cfg *tls.Config) Option {
	return func(c *Config) {
		c.TlsConfig = cfg
	}
}