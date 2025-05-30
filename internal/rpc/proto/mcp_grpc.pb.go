// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.29.3
// source: mcp.proto

package protopb

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	ModelContextService_GetModelContext_FullMethodName = "/rpc.proto.ModelContextService/GetModelContext"
)

// ModelContextServiceClient is the client API for ModelContextService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ModelContextServiceClient interface {
	// GetModelContext retrieves the runtime context for a given model
	GetModelContext(ctx context.Context, in *ModelContextRequest, opts ...grpc.CallOption) (*ModelContextResponse, error)
}

type modelContextServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewModelContextServiceClient(cc grpc.ClientConnInterface) ModelContextServiceClient {
	return &modelContextServiceClient{cc}
}

func (c *modelContextServiceClient) GetModelContext(ctx context.Context, in *ModelContextRequest, opts ...grpc.CallOption) (*ModelContextResponse, error) {
	out := new(ModelContextResponse)
	err := c.cc.Invoke(ctx, ModelContextService_GetModelContext_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ModelContextServiceServer is the server API for ModelContextService service.
// All implementations must embed UnimplementedModelContextServiceServer
// for forward compatibility
type ModelContextServiceServer interface {
	// GetModelContext retrieves the runtime context for a given model
	GetModelContext(context.Context, *ModelContextRequest) (*ModelContextResponse, error)
	mustEmbedUnimplementedModelContextServiceServer()
}

// UnimplementedModelContextServiceServer must be embedded to have forward compatible implementations.
type UnimplementedModelContextServiceServer struct {
}

func (UnimplementedModelContextServiceServer) GetModelContext(context.Context, *ModelContextRequest) (*ModelContextResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetModelContext not implemented")
}
func (UnimplementedModelContextServiceServer) mustEmbedUnimplementedModelContextServiceServer() {}

// UnsafeModelContextServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ModelContextServiceServer will
// result in compilation errors.
type UnsafeModelContextServiceServer interface {
	mustEmbedUnimplementedModelContextServiceServer()
}

func RegisterModelContextServiceServer(s grpc.ServiceRegistrar, srv ModelContextServiceServer) {
	s.RegisterService(&ModelContextService_ServiceDesc, srv)
}

func _ModelContextService_GetModelContext_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ModelContextRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ModelContextServiceServer).GetModelContext(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: ModelContextService_GetModelContext_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ModelContextServiceServer).GetModelContext(ctx, req.(*ModelContextRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ModelContextService_ServiceDesc is the grpc.ServiceDesc for ModelContextService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ModelContextService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "rpc.proto.ModelContextService",
	HandlerType: (*ModelContextServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetModelContext",
			Handler:    _ModelContextService_GetModelContext_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mcp.proto",
}
