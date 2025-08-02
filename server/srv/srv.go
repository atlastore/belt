package srv

import (
	"context"

	"github.com/atlastore/belt/server/options"
)

type Server interface {
	Start(ctx context.Context, addr string) error
	Close() error
	AddStopMonitoringFunc(fn options.StopMonitoringFunc)
}