package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atlastore/belt/logx"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

var (
	_ Server = &httpServer{}
)

type httpServer struct {
	log *logx.Logger
	app *fiber.App
	cfg config
	stopMonitors []StopMonitoringFunc
	ln net.Listener
}

func newHttpServer(log *logx.Logger, cfg config) *httpServer {
	app := fiber.New(cfg.fiberCfg)

	app.Use(recover.New())
	app.Use(loggingMiddleware(log))
	app.Use(compress.New(
		compress.Config{
			Level: compress.LevelBestCompression,
		},
	))

	app.Get("/healthz", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	if cfg.router != nil {
		cfg.router.RegisterRoutes(app)
	}

	return &httpServer{
		log: log,
		app: app,
		cfg: cfg,
		stopMonitors: cfg.stopMonitoring,
	}
}

func (hs *httpServer) Start(ctx context.Context, addr string) error {
	hs.print(addr)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	for _ ,fn := range hs.cfg.stopMonitoring {
		fn(ctx)
	}

	errChan := make(chan error)
	
	go func() {
		ln, err := net.Listen("tcp", addr)

		if err != nil {
			hs.log.Error("failed to bind listener", zap.String("id", hs.log.Config().InstanceID), zap.Error(err))
			errChan <- fmt.Errorf("failed to bind listener: %w", err)
			return
		}
		listenCfg := fiber.ListenConfig{
			DisableStartupMessage: true,
			GracefulContext: ctx,
		}

		if hs.cfg.tlsConfig != nil {
			ln = tls.NewListener(ln, hs.cfg.tlsConfig)
			hs.log.Info("serving HTTP with TLS", zap.String("id", hs.log.Config().InstanceID))
		}

		if err := hs.app.Listener(ln, listenCfg); err != nil {
			hs.log.Error("failed to start server", zap.String("id", hs.log.Config().InstanceID), zap.String("address", addr))
			errChan <- fmt.Errorf("server error: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		hs.log.Info("Shutting down servers")

		err := hs.Close()
		if err != nil {
			return  err
		}
		hs.log.Info("Servers shut down cleanly.")

		return  nil
	case err := <-errChan:
		return err
	}
}

func (hs *httpServer) Close() error {
	return hs.app.Shutdown()
}

func (hs *httpServer) AddStopMonitoringFunc(fn StopMonitoringFunc) {
	hs.stopMonitors = append(hs.stopMonitors, fn)
}

func (hs *httpServer) print(addr string) {
	fmt.Println("------------------------------------------------------------------------------------------------------")
	hs.log.Info(fmt.Sprintf("Listening on: http://%s ", addr))
	fmt.Println("------------------------------------------------------------------------------------------------------")
}

func loggingMiddleware(log *logx.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)

		status := c.Response().StatusCode()
		method := c.Method()
		path := c.OriginalURL()
		ip := c.IP()

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("ip", ip),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Error("HTTP request failed", fields...)
			// let Fiber handle it (e.g., custom error handler middleware)
			return err
		}

		log.Info("HTTP request", fields...)
		return nil
	}
}