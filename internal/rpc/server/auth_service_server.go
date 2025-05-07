package server

import (
	"context"

	pb "github.com/distributedJob/internal/rpc/proto"
	"github.com/distributedJob/internal/service"
	"github.com/distributedJob/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthServiceServer implements the Auth gRPC service
type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewAuthServiceServer creates a new Auth service server
func NewAuthServiceServer(authService service.AuthService) *AuthServiceServer {
	return &AuthServiceServer{
		authService: authService,
	}
}

// Authenticate handles user authentication requests
func (s *AuthServiceServer) Authenticate(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	// Validate request
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}

	// Authenticate user
	accessToken, _, user, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		logger.Error("Authentication failed", "error", err, "username", req.Username)
		return &pb.AuthResponse{
			Success: false,
			Message: "Authentication failed: " + err.Error(),
		}, nil
	}

	// Return successful response with token
	// Note: We only return the access token via gRPC, refresh tokens are only handled through HTTP cookies
	return &pb.AuthResponse{
		Token:   accessToken,
		UserId:  user.ID,
		Success: true,
		Message: "Authentication successful",
	}, nil
}

// ValidateToken validates a JWT token
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.TokenValidationResponse, error) {
	// Validate request
	if req.Token == "" {
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	// Validate token
	userID, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		logger.Debug("Token validation failed", "error", err)
		return &pb.TokenValidationResponse{
			Valid: false,
		}, nil
	}

	// Get user permissions
	permissions, err := s.authService.GetUserPermissions(userID)
	if err != nil {
		logger.Error("Failed to get user permissions", "error", err, "userID", userID)
		return &pb.TokenValidationResponse{
			Valid:  true,
			UserId: userID,
		}, nil
	}

	// Return validation response
	return &pb.TokenValidationResponse{
		Valid:       true,
		UserId:      userID,
		Permissions: permissions,
	}, nil
}

// GetUserPermissions gets permissions for a user
func (s *AuthServiceServer) GetUserPermissions(ctx context.Context, req *pb.UserRequest) (*pb.PermissionsResponse, error) {
	// Validate request
	if req.UserId <= 0 {
		return nil, status.Error(codes.InvalidArgument, "valid user_id is required")
	}

	// Get user permissions
	permissions, err := s.authService.GetUserPermissions(req.UserId)
	if err != nil {
		logger.Error("Failed to get user permissions", "error", err, "userID", req.UserId)
		return &pb.PermissionsResponse{
			Success: false,
		}, nil
	}

	// Return permissions
	return &pb.PermissionsResponse{
		Permissions: permissions,
		Success:     true,
	}, nil
}
