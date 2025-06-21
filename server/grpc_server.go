package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atlastore/belt/logx"
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
	_ Server = &grpcServer{}
)

type grpcServer struct {
	log *logx.Logger
	server *grpc.Server
	ln net.Listener
	cfg config
	stopMonitors []StopMonitoringFunc
	healthServer *health.Server
}

func newGrpcServer(log *logx.Logger, cfg config) *grpcServer {
	cfg.grpcOptions = append(cfg.grpcOptions, 
		grpc.UnaryInterceptor(unaryLoggingInterceptor(log)), 
		grpc.StreamInterceptor(streamLoggingInterceptor(log)),
	)

	if cfg.tlsConfig != nil {
		creds := credentials.NewTLS(cfg.tlsConfig)
		cfg.grpcOptions = append(cfg.grpcOptions, grpc.Creds(creds))
	}

	server := grpc.NewServer(cfg.grpcOptions...)

	hs := health.NewServer()
	grpc_health_v1.RegisterHealthServer(server, hs)
	hs.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	for _, reg := range cfg.registries {
		reg.Registrar(server, reg.Service)
	}

	if log.Config().State == logx.Development {
		reflection.Register(server)
	}

	return &grpcServer{
		log: log,
		server: server,
		healthServer: hs,
		cfg: cfg,
		stopMonitors: cfg.stopMonitoring,
	}
}

func (gs *grpcServer) Start(ctx context.Context, addr string) error {
	gs.print(addr)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	gs.ln = ln

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	for _ ,fn := range gs.cfg.stopMonitoring {
		fn(ctx)
	}

	errChan := make(chan error)

	go func() {
		if err := gs.server.Serve(ln); err != nil {
			gs.log.Error("failed to start server", zap.String("id", gs.log.Config().InstanceID), zap.String("address", addr))
			errChan <- fmt.Errorf("server error: %v", err)
		}
	}()

	select {
	case <-ctx.Done():
		gs.log.Info("Shutting down servers")

		err := gs.Close()
		if err != nil {
			return  err
		}
		gs.log.Info("Servers shut down cleanly.")

		return  nil
	case err := <-errChan:
		return err
	}
}

func (gs *grpcServer) Close() error {
	gs.SetServing(false)
	gs.server.GracefulStop()

	if gs.ln == nil {
		return nil
	}

	return gs.ln.Close()
}

func (gs *grpcServer) AddStopMonitoringFunc(fn StopMonitoringFunc) {
	gs.stopMonitors = append(gs.stopMonitors, fn)
}

func (gs *grpcServer) SetServing(isServing bool) {
	status := grpc_health_v1.HealthCheckResponse_NOT_SERVING
	if isServing {
		status = grpc_health_v1.HealthCheckResponse_SERVING
	}
	gs.healthServer.SetServingStatus("", status)
}
func (gs *grpcServer) print(addr string) {
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