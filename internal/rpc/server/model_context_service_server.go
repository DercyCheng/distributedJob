package server

import (
	"context"

	pb "distributedJob/internal/rpc/proto"
	"distributedJob/internal/service"
)

// mcpServiceServer implements the ModelContextServiceServer interface
type mcpServiceServer struct {
	pb.UnimplementedModelContextServiceServer
	service service.ModelContextService
}

// NewModelContextServiceServer creates a new ModelContextServiceServer
func NewModelContextServiceServer(svc service.ModelContextService) pb.ModelContextServiceServer {
	return &mcpServiceServer{service: svc}
}

// GetModelContext handles the GetModelContext RPC
func (s *mcpServiceServer) GetModelContext(ctx context.Context, req *pb.ModelContextRequest) (*pb.ModelContextResponse, error) {
	ctxStr, err := s.service.GetModelContext(req.ModelId)
	if err != nil {
		return nil, err
	}
	return &pb.ModelContextResponse{Context: ctxStr}, nil
}
