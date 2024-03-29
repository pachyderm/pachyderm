// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.25.1
// source: logs/logs.proto

package logs

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
	API_GetLogs_FullMethodName = "/logs.API/GetLogs"
)

// APIClient is the client API for API service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type APIClient interface {
	GetLogs(ctx context.Context, in *GetLogsRequest, opts ...grpc.CallOption) (API_GetLogsClient, error)
}

type aPIClient struct {
	cc grpc.ClientConnInterface
}

func NewAPIClient(cc grpc.ClientConnInterface) APIClient {
	return &aPIClient{cc}
}

func (c *aPIClient) GetLogs(ctx context.Context, in *GetLogsRequest, opts ...grpc.CallOption) (API_GetLogsClient, error) {
	stream, err := c.cc.NewStream(ctx, &API_ServiceDesc.Streams[0], API_GetLogs_FullMethodName, opts...)
	if err != nil {
		return nil, err
	}
	x := &aPIGetLogsClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type API_GetLogsClient interface {
	Recv() (*GetLogsResponse, error)
	grpc.ClientStream
}

type aPIGetLogsClient struct {
	grpc.ClientStream
}

func (x *aPIGetLogsClient) Recv() (*GetLogsResponse, error) {
	m := new(GetLogsResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// APIServer is the server API for API service.
// All implementations must embed UnimplementedAPIServer
// for forward compatibility
type APIServer interface {
	GetLogs(*GetLogsRequest, API_GetLogsServer) error
	mustEmbedUnimplementedAPIServer()
}

// UnimplementedAPIServer must be embedded to have forward compatible implementations.
type UnimplementedAPIServer struct {
}

func (UnimplementedAPIServer) GetLogs(*GetLogsRequest, API_GetLogsServer) error {
	return status.Errorf(codes.Unimplemented, "method GetLogs not implemented")
}
func (UnimplementedAPIServer) mustEmbedUnimplementedAPIServer() {}

// UnsafeAPIServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to APIServer will
// result in compilation errors.
type UnsafeAPIServer interface {
	mustEmbedUnimplementedAPIServer()
}

func RegisterAPIServer(s grpc.ServiceRegistrar, srv APIServer) {
	s.RegisterService(&API_ServiceDesc, srv)
}

func _API_GetLogs_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(GetLogsRequest)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(APIServer).GetLogs(m, &aPIGetLogsServer{stream})
}

type API_GetLogsServer interface {
	Send(*GetLogsResponse) error
	grpc.ServerStream
}

type aPIGetLogsServer struct {
	grpc.ServerStream
}

func (x *aPIGetLogsServer) Send(m *GetLogsResponse) error {
	return x.ServerStream.SendMsg(m)
}

// API_ServiceDesc is the grpc.ServiceDesc for API service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var API_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "logs.API",
	HandlerType: (*APIServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "GetLogs",
			Handler:       _API_GetLogs_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "logs/logs.proto",
}
