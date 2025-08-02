package grpc_server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/atlastore/belt/logx"
	"github.com/atlastore/belt/server/options"
	"github.com/atlastore/belt/server/srv"
	server_utils "github.com/atlastore/belt/server/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var (
	_ srv.Server = &GrpcServer{}
)

type GrpcServer struct {
	log *logx.Logger
	server *grpc.Server
	ln net.Listener
	cfg options.Config
	stopMonitors []options.StopMonitoringFunc
	healthServer *health.Server
	applyMonitor bool
}

func New(log *logx.Logger, cfg options.Config, applyMonitor bool) *GrpcServer {
	cfg.GrpcOptions = append(cfg.GrpcOptions, 
		grpc.UnaryInterceptor(unaryLoggingInterceptor(log)), 
		grpc.StreamInterceptor(streamLoggingInterceptor(log)),
	)

	if cfg.TlsConfig != nil {
		creds := credentials.NewTLS(cfg.TlsConfig)
		cfg.GrpcOptions = append(cfg.GrpcOptions, grpc.Creds(creds))
	}

	server := grpc.NewServer(cfg.GrpcOptions...)

	hs := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, hs)
	hs.SetServingStatus(log.Config().Service, grpc_health_v1.HealthCheckResponse_SERVING)

	for _, reg := range cfg.Registries {
		reg.Registrar(server, reg.Service)
	}

	if log.Config().State == logx.Development {
		reflection.Register(server)
	}

	return &GrpcServer{
		log: log,
		server: server,
		healthServer: hs,
		cfg: cfg,
		stopMonitors: cfg.StopMonitoring,
		applyMonitor: applyMonitor,
	}
}

func (gs *GrpcServer) Start(ctx context.Context, addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return gs.StartWithListener(ctx, ln)
}

func (gs *GrpcServer) StartWithListener(ctx context.Context, ln net.Listener) error {
	gs.print(ln.Addr().String())
	gs.ln = ln

	ctx, stop := server_utils.SignalContext(ctx)
	defer stop()

	if gs.applyMonitor {
		for _, fn := range gs.stopMonitors {
			go fn(ctx)
		}
	}

	errChan := make(chan error, 1)

	go func() {
		err := gs.server.Serve(ln)
		if err != nil && !strings.Contains(err.Error(), "listener closed") {
			gs.log.Error("gRPC server failed", zap.String("id", gs.log.Config().InstanceID), zap.String("address", ln.Addr().String()), zap.Error(err))
			errChan <- err
		}
	}()

	var once sync.Once
	var closeErr error

	doClose := func() {
		if err := gs.Close(); err != nil {
			closeErr = err
		}
	}

	select {
	case <-ctx.Done():
		if !gs.applyMonitor {
			return nil
		}
		gs.log.Debug("gRPC shutdown initiated")
		once.Do(doClose)
		gs.log.Info("gRPC shutdown cleanly")

		return closeErr
	case err := <-errChan:
		once.Do(doClose)
		if closeErr != nil {
			return errors.Join(closeErr, err)
		}
		return err
	}
}

func (gs *GrpcServer) Close() error {
	gs.SetServing(false)
	gs.server.GracefulStop()

	if gs.ln == nil {
		return nil
	}

	return nil
}

func (gs *GrpcServer) AddStopMonitoringFunc(fn options.StopMonitoringFunc) {
	gs.stopMonitors = append(gs.stopMonitors, fn)
}

func (gs *GrpcServer) SetServing(isServing bool) {
	status := grpc_health_v1.HealthCheckResponse_NOT_SERVING
	if isServing {
		status = grpc_health_v1.HealthCheckResponse_SERVING
	}
	gs.healthServer.SetServingStatus(gs.log.Config().Service, status)
}
func (gs *GrpcServer) print(addr string) {
	fmt.Println("------------------------------------------------------------------------------------------------------")
	gs.log.Info(fmt.Sprintf("Listening on: %s", addr))
	fmt.Println("------------------------------------------------------------------------------------------------------")
}

func unaryLoggingInterceptor(log *logx.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		st, _ := status.FromError(err)
		code := st.Code()
		peerInfo, _ := peer.FromContext(ctx)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("status", code.String()),
			zap.Duration("duration", duration),
			zap.String("peer", peerInfo.Addr.String()),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Error("gRPC request failed", fields...)
		} else {
			log.Info("gRPC request succeeded", fields...)
		}

		return resp, err
	}
}

func streamLoggingInterceptor(log *logx.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		start := time.Now()
		err := handler(srv, ss)
		duration := time.Since(start)

		st, _ := status.FromError(err)
		code := st.Code()
		peerInfo, _ := peer.FromContext(ss.Context())

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.String("status", code.String()),
			zap.Duration("duration", duration),
			zap.String("peer", peerInfo.Addr.String()),
		}

		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Error("gRPC stream request failed", fields...)
		} else {
			log.Info("gRPC stream request succeeded", fields...)
		}

		return err
	}
}