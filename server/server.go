package server

import (
	"context"

	"github.com/atlastore/belt/logx"
)

type Type int

const (
	GRPC Type = iota
	HTTP
)

func (t Type) String() string {
	switch t {
	case GRPC:
		return "gRPC"
	case HTTP:
		return "http"
	default:
		return ""
	}
}

type Server interface {
	Start(ctx context.Context, addr string) error
	Close() error
	AddStopMonitoringFunc(fn StopMonitoringFunc)
}

func NewServer(t Type, log *logx.Logger, options ...Option) Server {
	cfg := newConfig(options)
	switch t {
	case GRPC:
		return newGrpcServer(log, cfg)
	case HTTP:
		return newHttpServer(log, cfg)
	default:
		return nil
	}
}

