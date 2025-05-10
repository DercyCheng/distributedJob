package server

import (
	"context"
	"fmt"
	"net"
	"time"

	"distributedJob/internal/config"
	"distributedJobob"
	pb "distributedJobpc/proto"
	"distributedJobervice"
	"distributedJob"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// RPCServer represents the gRPC server
type RPCServer struct {
	config      *config.Config
	server      *grpc.Server
	taskService service.TaskService
	authService service.AuthService
	scheduler   *job.Scheduler
}

// NewRPCServer creates a new RPCServer instance
func NewRPCServer(
	config *config.Config,
	scheduler *job.Scheduler,
	taskService service.TaskService,
	authService service.AuthService,
) *RPCServer {
	// Configure keep-alive parameters
	kaParams := keepalive.ServerParameters{
		MaxConnectionIdle:     time.Duration(config.Rpc.KeepAliveTime) * time.Second,
		MaxConnectionAge:      time.Duration(config.Rpc.KeepAliveTime*3) * time.Second,
		MaxConnectionAgeGrace: time.Duration(config.Rpc.KeepAliveTimeout) * time.Second,
		Time:                  time.Duration(config.Rpc.KeepAliveTime) * time.Second,
		Timeout:               time.Duration(config.Rpc.KeepAliveTimeout) * time.Second,
	}

	// Create gRPC server with keep-alive and stream interceptors
	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(kaParams),
		grpc.MaxConcurrentStreams(uint32(config.Rpc.MaxConcurrentStreams)),
		grpc.Creds(insecure.NewCredentials()),
	)

	return &RPCServer{
		config:      config,
		server:      grpcServer,
		taskService: taskService,
		authService: authService,
		scheduler:   scheduler,
	}
}

// Start initializes and starts the RPC server
func (s *RPCServer) Start() error {
	// Register services
	s.registerServices()

	// Register reflection service for gRPC CLI
	reflection.Register(s.server)

	// Start listening
	addr := fmt.Sprintf(":%d", s.config.Rpc.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", addr, err)
	}

	logger.Infof("RPC server listening on %s", addr)

	// Start serving (blocking call)
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}

	return nil
}

// StartAsync starts the RPC server in a goroutine
func (s *RPCServer) StartAsync() {
	go func() {
		if err := s.Start(); err != nil {
			logger.Errorf("RPC server error: %v", err)
		}
	}()
}

// Stop gracefully stops the RPC server
func (s *RPCServer) Stop() {
	if s.server != nil {
		logger.Info("Gracefully stopping RPC server...")
		s.server.GracefulStop()
		logger.Info("RPC server stopped")
	}
}

// registerServices registers all gRPC services
func (s *RPCServer) registerServices() {
	// Register the task scheduler service
	taskSchedulerServer := NewTaskSchedulerServer(s.scheduler)
	pb.RegisterTaskSchedulerServer(s.server, taskSchedulerServer)

	// Register the auth service
	authServiceServer := NewAuthServiceServer(s.authService)
	pb.RegisterAuthServiceServer(s.server, authServiceServer)

	// Register the data service
	dataServiceServer := NewDataServiceServer(s.taskService)
	pb.RegisterDataServiceServer(s.server, dataServiceServer)

	logger.Info("Registered all RPC services")
}

// GetClientConnection returns a gRPC client connection to this server
func GetClientConnection(target string, timeout time.Duration) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Configure dial options
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                10 * time.Second, // Send pings every 10 seconds if there is no activity
			Timeout:             3 * time.Second,  // Wait 3 seconds for ping ack before considering the connection dead
			PermitWithoutStream: true,             // Send pings even without active streams
		}),
	}

	// Connect to the server
	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", target, err)
	}

	return conn, nil
}
