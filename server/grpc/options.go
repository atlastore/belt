package grpc

import (
	"github.com/atlastore/belt/server/options"
	"google.golang.org/grpc"
)

func WithConfig(cfg ...grpc.ServerOption) options.Option {
	return func(c *options.Config) {
		c.GrpcOptions = cfg
	}
}

func WithRegistry[T any](service T, registrar options.RegisterFunc[T]) options.Option {
	wrappedRegistrar := func (s grpc.ServiceRegistrar, _ any)  {
		registrar(s, service)
	}

	return func(c *options.Config) {
		c.Registries = append(c.Registries, options.ServiceRegistry[any]{
			Service: service,
			Registrar: wrappedRegistrar,
		})
	}
}

func WithRegistries(registries ...options.Option) options.Option {
	return func(c *options.Config) {
		for _, opt := range registries {
			opt(c)
		}
	}
}