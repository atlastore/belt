package http_server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/atlastore/belt/logx"
	"github.com/atlastore/belt/server/options"
	"github.com/atlastore/belt/server/srv"
	server_utils "github.com/atlastore/belt/server/utils"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"go.uber.org/zap"
)

var (
	_ srv.Server = &HttpServer{}
)

type HttpServer struct {
	log *logx.Logger
	app *fiber.App
	cfg options.Config
	stopMonitors []options.StopMonitoringFunc
	ln net.Listener
	applyMonitor bool
}

func New(log *logx.Logger, cfg options.Config, applyMonitor bool) *HttpServer {
	app := fiber.New(cfg.FiberCfg)

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

	if cfg.Router != nil {
		cfg.Router.RegisterRoutes(app)
	}

	return &HttpServer{
		log: log,
		app: app,
		cfg: cfg,
		stopMonitors: cfg.StopMonitoring,
		applyMonitor: applyMonitor,
	}
}

func (hs *HttpServer) Start(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return hs.StartWithListener(ctx, ln)
}

func (hs *HttpServer) StartWithListener(ctx context.Context, ln net.Listener) error {
	hs.print(ln.Addr().String())

	ctx, stop := server_utils.SignalContext(ctx)
	defer stop()

	if hs.applyMonitor {
		for _, fn := range hs.stopMonitors {
			go fn(ctx)
		}
	}

	errChan := make(chan error, 1)

	go func() {
		listenCfg := fiber.ListenConfig{
			DisableStartupMessage: true,
			GracefulContext: ctx,
		}

		if hs.cfg.TlsConfig != nil {
			ln = tls.NewListener(ln, hs.cfg.TlsConfig)
			hs.log.Info("serving HTTP with TLS")
		}

		err := hs.app.Listener(ln, listenCfg)
		if err != nil && !strings.Contains(err.Error(), "listener closed") {
			hs.log.Error("failed to start HTTP server", zap.String("address", ln.Addr().String()), zap.Error(err))
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		if !hs.applyMonitor {
			return nil
		}
		hs.log.Debug("HTTP shutdown initiated")
		err := hs.Close()
		if err != nil {
			return err
		}
		hs.log.Info("HTTP shutdown cleanly")

		return nil
	case err := <-errChan:
		err2 := hs.Close()
		if err2 != nil {
			return errors.Join(err, err2)
		}
		return err
	}
}

func (hs *HttpServer) Close() error {
	return hs.app.Shutdown()
}

func (hs *HttpServer) AddStopMonitoringFunc(fn options.StopMonitoringFunc) {
	hs.stopMonitors = append(hs.stopMonitors, fn)
}

func (hs *HttpServer) print(addr string) {
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