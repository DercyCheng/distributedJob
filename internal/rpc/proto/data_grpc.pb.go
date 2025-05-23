// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.29.1
// source: internal/rpc/proto/data.proto

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
	DataService_GetTaskHistory_FullMethodName = "/data.DataService/GetTaskHistory"
	DataService_GetStatistics_FullMethodName  = "/data.DataService/GetStatistics"
)

// DataServiceClient is the client API for DataService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DataServiceClient interface {
	GetTaskHistory(ctx context.Context, in *TaskHistoryRequest, opts ...grpc.CallOption) (*TaskHistoryResponse, error)
	GetStatistics(ctx context.Context, in *StatisticsRequest, opts ...grpc.CallOption) (*StatisticsResponse, error)
}

type dataServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewDataServiceClient(cc grpc.ClientConnInterface) DataServiceClient {
	return &dataServiceClient{cc}
}

func (c *dataServiceClient) GetTaskHistory(ctx context.Context, in *TaskHistoryRequest, opts ...grpc.CallOption) (*TaskHistoryResponse, error) {
	out := new(TaskHistoryResponse)
	err := c.cc.Invoke(ctx, DataService_GetTaskHistory_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *dataServiceClient) GetStatistics(ctx context.Context, in *StatisticsRequest, opts ...grpc.CallOption) (*StatisticsResponse, error) {
	out := new(StatisticsResponse)
	err := c.cc.Invoke(ctx, DataService_GetStatistics_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DataServiceServer is the server API for DataService service.
// All implementations must embed UnimplementedDataServiceServer
// for forward compatibility
type DataServiceServer interface {
	GetTaskHistory(context.Context, *TaskHistoryRequest) (*TaskHistoryResponse, error)
	GetStatistics(context.Context, *StatisticsRequest) (*StatisticsResponse, error)
	mustEmbedUnimplementedDataServiceServer()
}

// UnimplementedDataServiceServer must be embedded to have forward compatible implementations.
type UnimplementedDataServiceServer struct {
}

func (UnimplementedDataServiceServer) GetTaskHistory(context.Context, *TaskHistoryRequest) (*TaskHistoryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTaskHistory not implemented")
}
func (UnimplementedDataServiceServer) GetStatistics(context.Context, *StatisticsRequest) (*StatisticsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStatistics not implemented")
}
func (UnimplementedDataServiceServer) mustEmbedUnimplementedDataServiceServer() {}

// UnsafeDataServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DataServiceServer will
// result in compilation errors.
type UnsafeDataServiceServer interface {
	mustEmbedUnimplementedDataServiceServer()
}

func RegisterDataServiceServer(s grpc.ServiceRegistrar, srv DataServiceServer) {
	s.RegisterService(&DataService_ServiceDesc, srv)
}

func _DataService_GetTaskHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TaskHistoryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataServiceServer).GetTaskHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataService_GetTaskHistory_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataServiceServer).GetTaskHistory(ctx, req.(*TaskHistoryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _DataService_GetStatistics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatisticsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DataServiceServer).GetStatistics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: DataService_GetStatistics_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DataServiceServer).GetStatistics(ctx, req.(*StatisticsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// DataService_ServiceDesc is the grpc.ServiceDesc for DataService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var DataService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "data.DataService",
	HandlerType: (*DataServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetTaskHistory",
			Handler:    _DataService_GetTaskHistory_Handler,
		},
		{
			MethodName: "GetStatistics",
			Handler:    _DataService_GetStatistics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "internal/rpc/proto/data.proto",
}
