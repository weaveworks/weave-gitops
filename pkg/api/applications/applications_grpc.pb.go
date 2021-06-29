// Code generated by protoc-gen-go-grpc. DO NOT EDIT.

package applications

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

// ApplicationsClient is the client API for Applications service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ApplicationsClient interface {
	//
	// ListApplications returns the list of WeGo applications that the authenticated user has access to.
	ListApplications(ctx context.Context, in *ListApplicationsRequest, opts ...grpc.CallOption) (*ListApplicationsResponse, error)
	//
	// GetApplication returns a given application
	GetApplication(ctx context.Context, in *GetApplicationRequest, opts ...grpc.CallOption) (*GetApplicationResponse, error)
}

type applicationsClient struct {
	cc grpc.ClientConnInterface
}

func NewApplicationsClient(cc grpc.ClientConnInterface) ApplicationsClient {
	return &applicationsClient{cc}
}

func (c *applicationsClient) ListApplications(ctx context.Context, in *ListApplicationsRequest, opts ...grpc.CallOption) (*ListApplicationsResponse, error) {
	out := new(ListApplicationsResponse)
	err := c.cc.Invoke(ctx, "/wego_server.v1.Applications/ListApplications", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *applicationsClient) GetApplication(ctx context.Context, in *GetApplicationRequest, opts ...grpc.CallOption) (*GetApplicationResponse, error) {
	out := new(GetApplicationResponse)
	err := c.cc.Invoke(ctx, "/wego_server.v1.Applications/GetApplication", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ApplicationsServer is the server API for Applications service.
// All implementations must embed UnimplementedApplicationsServer
// for forward compatibility
type ApplicationsServer interface {
	//
	// ListApplications returns the list of WeGo applications that the authenticated user has access to.
	ListApplications(context.Context, *ListApplicationsRequest) (*ListApplicationsResponse, error)
	//
	// GetApplication returns a given application
	GetApplication(context.Context, *GetApplicationRequest) (*GetApplicationResponse, error)
	mustEmbedUnimplementedApplicationsServer()
}

// UnimplementedApplicationsServer must be embedded to have forward compatible implementations.
type UnimplementedApplicationsServer struct {
}

func (UnimplementedApplicationsServer) ListApplications(context.Context, *ListApplicationsRequest) (*ListApplicationsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ListApplications not implemented")
}
func (UnimplementedApplicationsServer) GetApplication(context.Context, *GetApplicationRequest) (*GetApplicationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetApplication not implemented")
}
func (UnimplementedApplicationsServer) mustEmbedUnimplementedApplicationsServer() {}

// UnsafeApplicationsServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ApplicationsServer will
// result in compilation errors.
type UnsafeApplicationsServer interface {
	mustEmbedUnimplementedApplicationsServer()
}

func RegisterApplicationsServer(s grpc.ServiceRegistrar, srv ApplicationsServer) {
	s.RegisterService(&Applications_ServiceDesc, srv)
}

func _Applications_ListApplications_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ListApplicationsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationsServer).ListApplications(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/wego_server.v1.Applications/ListApplications",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationsServer).ListApplications(ctx, req.(*ListApplicationsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Applications_GetApplication_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetApplicationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ApplicationsServer).GetApplication(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/wego_server.v1.Applications/GetApplication",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ApplicationsServer).GetApplication(ctx, req.(*GetApplicationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Applications_ServiceDesc is the grpc.ServiceDesc for Applications service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Applications_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "wego_server.v1.Applications",
	HandlerType: (*ApplicationsServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ListApplications",
			Handler:    _Applications_ListApplications_Handler,
		},
		{
			MethodName: "GetApplication",
			Handler:    _Applications_GetApplication_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "api/applications/applications.proto",
}
