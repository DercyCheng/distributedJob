package server

import (
	"context"
	
	"github.com/distributedJob/internal/service"
	pb "github.com/distributedJob/internal/rpc/proto"
	"github.com/distributedJob/pkg/logger"
)

// AuthServiceServer implements the AuthService RPC service
type AuthServiceServer struct {
	pb.UnimplementedAuthServiceServer
	authService service.AuthService
}

// NewAuthServiceServer creates a new AuthServiceServer
func NewAuthServiceServer(authService service.AuthService) *AuthServiceServer {
	return &AuthServiceServer{authService: authService}
}

// Authenticate implements the Authenticate RPC method
func (s *AuthServiceServer) Authenticate(ctx context.Context, req *pb.AuthRequest) (*pb.AuthResponse, error) {
	token, err := s.authService.Login(req.Username, req.Password)
	if err != nil {
		logger.Errorf("Failed to authenticate user via RPC: %v", err)
		return &pb.AuthResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}
	
	// For the response, we need a userID but don't have it from the interface
	// We can try to get it from the token validation
	claims, err := s.authService.ValidateToken(token)
	var userID int64
	if err == nil && claims != nil {
		userID = claims.UserID
	}
	
	return &pb.AuthResponse{
		Token:   token,
		UserId:  userID, // Keep as int64
		Success: true,
		Message: "Authentication successful",
	}, nil
}

// ValidateToken implements the ValidateToken RPC method
func (s *AuthServiceServer) ValidateToken(ctx context.Context, req *pb.TokenRequest) (*pb.TokenValidationResponse, error) {
	claims, err := s.authService.ValidateToken(req.Token)
	if err != nil {
		return &pb.TokenValidationResponse{
			Valid: false,
		}, nil
	}
	
	// Since Claims contains UserID, we can access it directly
	return &pb.TokenValidationResponse{
		Valid:       true,
		UserId:      claims.UserID, // Keep as int64
		// We would need to get permissions separately, but since we don't have a direct method
		// in the service interface, we'll leave this empty for now
		Permissions: []string{},
	}, nil
}

// GetUserPermissions implements the GetUserPermissions RPC method
func (s *AuthServiceServer) GetUserPermissions(ctx context.Context, req *pb.UserRequest) (*pb.PermissionsResponse, error) {
	// Since there's no direct method to get all permissions for a user in the AuthService interface,
	// we'll need to use other methods to synthesize this functionality
	
	// First get all available permissions
	permissionList, err := s.authService.GetPermissionList()
	if err != nil {
		logger.Errorf("Failed to get permission list via RPC: %v", err)
		return &pb.PermissionsResponse{
			Success: false,
		}, nil
	}
	
	// Use userID directly without parsing
	userID := req.UserId
	
	var userPermissions []string
	for _, perm := range permissionList {
		hasPermission, err := s.authService.HasPermission(userID, perm.Code)
		if err == nil && hasPermission {
			userPermissions = append(userPermissions, perm.Code)
		}
	}
	
	return &pb.PermissionsResponse{
		Permissions: userPermissions,
		Success:     true,
	}, nil
}