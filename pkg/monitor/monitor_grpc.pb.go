// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.22.0
// source: pkg/monitor/monitor.proto

package monitor

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

// MonitorClient is the client API for Monitor service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MonitorClient interface {
	SendMetrics(ctx context.Context, in *SendMetricsRequest, opts ...grpc.CallOption) (*SendMetricsResponse, error)
}

type monitorClient struct {
	cc grpc.ClientConnInterface
}

func NewMonitorClient(cc grpc.ClientConnInterface) MonitorClient {
	return &monitorClient{cc}
}

func (c *monitorClient) SendMetrics(ctx context.Context, in *SendMetricsRequest, opts ...grpc.CallOption) (*SendMetricsResponse, error) {
	out := new(SendMetricsResponse)
	err := c.cc.Invoke(ctx, "/Monitor/SendMetrics", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MonitorServer is the server API for Monitor service.
// All implementations must embed UnimplementedMonitorServer
// for forward compatibility
type MonitorServer interface {
	SendMetrics(context.Context, *SendMetricsRequest) (*SendMetricsResponse, error)
	mustEmbedUnimplementedMonitorServer()
}

// UnimplementedMonitorServer must be embedded to have forward compatible implementations.
type UnimplementedMonitorServer struct {
}

func (UnimplementedMonitorServer) SendMetrics(context.Context, *SendMetricsRequest) (*SendMetricsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMetrics not implemented")
}
func (UnimplementedMonitorServer) mustEmbedUnimplementedMonitorServer() {}

// UnsafeMonitorServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MonitorServer will
// result in compilation errors.
type UnsafeMonitorServer interface {
	mustEmbedUnimplementedMonitorServer()
}

func RegisterMonitorServer(s grpc.ServiceRegistrar, srv MonitorServer) {
	s.RegisterService(&Monitor_ServiceDesc, srv)
}

func _Monitor_SendMetrics_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMetricsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MonitorServer).SendMetrics(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Monitor/SendMetrics",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MonitorServer).SendMetrics(ctx, req.(*SendMetricsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Monitor_ServiceDesc is the grpc.ServiceDesc for Monitor service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Monitor_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "Monitor",
	HandlerType: (*MonitorServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMetrics",
			Handler:    _Monitor_SendMetrics_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "pkg/monitor/monitor.proto",
}
