package mux_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/atlastore/belt/logx"
	"github.com/atlastore/belt/server/grpc"
	"github.com/atlastore/belt/server/http"
	"github.com/atlastore/belt/server/options"
	"github.com/atlastore/belt/server/srv"
	server_utils "github.com/atlastore/belt/server/utils"
	"github.com/cockroachdb/cmux"
	"go.uber.org/zap"
)

var (
	_ srv.Server = &muxServer{}
)

type muxServer struct {
	log *logx.Logger
	cfg options.Config
	grpcS *grpc.Server
	httpS *http.Server
	m cmux.CMux
	ln net.Listener
}

func New(log *logx.Logger, cfg options.Config) *muxServer {
	return &muxServer{
		log: log,
		cfg: cfg,
	}
}

type nonClosingListener struct {
	net.Listener
}

func (ncl nonClosingListener) Close() error {
	// Override Close to no-op
	return nil
}

func (m *muxServer) Start(ctx context.Context, addr string) error {
	var err error
	m.ln, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("mux: failed to listen on %s: %w", addr, err)
	}

	m.m = cmux.New(m.ln)
	grpcL := m.m.Match(cmux.HTTP2())
	httpL := m.m.Match(cmux.HTTP1Fast())
	
	safeGrpcL := nonClosingListener{grpcL}
	safeHttpL := nonClosingListener{httpL}

	grpcLog := m.log.With(zap.String("component", "grpc"))
	httpLog := m.log.With(zap.String("component", "http"))

	m.grpcS = grpc.New(grpcLog, m.cfg, false)
	m.httpS = http.New(httpLog, m.cfg, false)

	for _, fn := range m.cfg.StopMonitoring {
		go fn(ctx)
	}

	ctx, stop := server_utils.SignalContext(ctx)
	defer stop()

	errChan := make(chan error, 3)


	go func() {
		err := m.grpcS.StartWithListener(ctx, safeGrpcL)
		if err != nil {
			m.log.Error("failed to start mux gRPC server", zap.String("address", addr))
			errChan <- err
		}
	}()

	go func() {
		err := m.httpS.StartWithListener(ctx, safeHttpL)
		if err != nil {
			m.log.Error("failed to start mux HTTP server", zap.String("address", addr))
			errChan <- err
		}
	}()

	go func () {
		err := m.m.Serve()
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Println(err)
			m.log.Error("failed to start mux server", zap.String("address", addr))
			errChan <- err
		}
	}()

	var once sync.Once
	var closeErr error

	doClose := func() {
		if err := m.Close(); err != nil {
			closeErr = err
		}
	}

	select {
	case <-ctx.Done():
		once.Do(doClose)
		return closeErr
	case err := <-errChan:
		once.Do(doClose)
		if closeErr != nil {
			return errors.Join(closeErr, err)
		}
		return err
	}
}

func (m *muxServer) Close() error {
	m.log.Debug("Mux shutdown initiated")
	if m.ln != nil {
		if err := m.ln.Close(); err != nil {
			return err
		}
	}
	m.log.Info("Mux shutdown cleanly")
	return nil
}

func (m *muxServer) AddStopMonitoringFunc(fn options.StopMonitoringFunc) {
	m.cfg.StopMonitoring = append(m.cfg.StopMonitoring, fn)
}