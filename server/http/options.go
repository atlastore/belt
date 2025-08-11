package http

import (
	"github.com/atlastore/belt/server/options"
	"github.com/gofiber/fiber/v3"
)

func WithConfig(cfg fiber.Config) options.Option {
	return func(c *options.Config) {
		c.FiberCfg = cfg
	}
}

func WithRouter(r options.Router) options.Option {
	return func(c *options.Config) {
		c.Router = r
	}
}