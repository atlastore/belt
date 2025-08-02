package server

import (
	"context"

	"github.com/atlastore/belt/logx"
	"github.com/atlastore/belt/server/grpc_server"
	"github.com/atlastore/belt/server/http_server"
	"github.com/atlastore/belt/server/mux_server"
	"github.com/atlastore/belt/server/options"
	"go.uber.org/zap"
)

type Type int

const (
	GRPC Type = iota
	HTTP
	MUX
)

func (t Type) String() string {
	switch t {
	case GRPC:
		return "gRPC"
	case HTTP:
		return "http"
	case MUX:
		return "mux"
	default:
		return ""
	}
}

type Server interface {
	Start(ctx context.Context, addr string) error
	Close() error
	AddStopMonitoringFunc(fn options.StopMonitoringFunc)
}

func NewServer(t Type, log *logx.Logger, opts ...options.Option) Server {
	cfg := options.NewConfig(opts)
	switch t {
	case GRPC:
		return grpc_server.New(log, cfg, true)
	case HTTP:
		return http_server.New(log, cfg, true)
	case MUX:
		return mux_server.New(log, cfg)
	default:
		log.Fatal("could not start server type is unknown", zap.String("type", t.String()))
		return nil
	}
}

