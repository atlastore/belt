package server_utils

import (
	"context"
	"os/signal"
	"syscall"
)


func SignalContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		return context.WithCancel(context.Background())
	}
	
	return signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
}
