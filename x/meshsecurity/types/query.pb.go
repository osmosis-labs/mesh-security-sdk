// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/meshsecurity/v1beta1/query.proto

package types

import (
	context "context"
	fmt "fmt"
	io "io"
	math "math"
	math_bits "math/bits"

	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = proto.Marshal
	_ = fmt.Errorf
	_ = math.Inf
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// QueryVirtualStakingMaxCapLimitRequest is the request type for the
// Query/VirtualStakingMaxCapLimit RPC method
type QueryVirtualStakingMaxCapLimitRequest struct {
	// Address is the address of the contract to query
	Address string `protobuf:"bytes,1,opt,name=address,proto3" json:"address,omitempty"`
}

func (m *QueryVirtualStakingMaxCapLimitRequest) Reset()         { *m = QueryVirtualStakingMaxCapLimitRequest{} }
func (m *QueryVirtualStakingMaxCapLimitRequest) String() string { return proto.CompactTextString(m) }
func (*QueryVirtualStakingMaxCapLimitRequest) ProtoMessage()    {}
func (*QueryVirtualStakingMaxCapLimitRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_50c89ba006eed4fb, []int{0}
}

func (m *QueryVirtualStakingMaxCapLimitRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *QueryVirtualStakingMaxCapLimitRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryVirtualStakingMaxCapLimitRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *QueryVirtualStakingMaxCapLimitRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryVirtualStakingMaxCapLimitRequest.Merge(m, src)
}

func (m *QueryVirtualStakingMaxCapLimitRequest) XXX_Size() int {
	return m.Size()
}

func (m *QueryVirtualStakingMaxCapLimitRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryVirtualStakingMaxCapLimitRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryVirtualStakingMaxCapLimitRequest proto.InternalMessageInfo

// QueryVirtualStakingMaxCapLimitResponse is the response type for the
// Query/VirtualStakingMaxCapLimit RPC method
type QueryVirtualStakingMaxCapLimitResponse struct {
	Delegated types.Coin `protobuf:"bytes,1,opt,name=delegated,proto3" json:"delegated"`
	Cap       types.Coin `protobuf:"bytes,2,opt,name=cap,proto3" json:"cap"`
}

func (m *QueryVirtualStakingMaxCapLimitResponse) Reset() {
	*m = QueryVirtualStakingMaxCapLimitResponse{}
}
func (m *QueryVirtualStakingMaxCapLimitResponse) String() string { return proto.CompactTextString(m) }
func (*QueryVirtualStakingMaxCapLimitResponse) ProtoMessage()    {}
func (*QueryVirtualStakingMaxCapLimitResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_50c89ba006eed4fb, []int{1}
}

func (m *QueryVirtualStakingMaxCapLimitResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *QueryVirtualStakingMaxCapLimitResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryVirtualStakingMaxCapLimitResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *QueryVirtualStakingMaxCapLimitResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryVirtualStakingMaxCapLimitResponse.Merge(m, src)
}

func (m *QueryVirtualStakingMaxCapLimitResponse) XXX_Size() int {
	return m.Size()
}

func (m *QueryVirtualStakingMaxCapLimitResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryVirtualStakingMaxCapLimitResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryVirtualStakingMaxCapLimitResponse proto.InternalMessageInfo

// QueryVirtualStakingMaxCapLimitsRequest is the request type for the
// Query/VirtualStakingMaxCapLimits RPC method
type QueryVirtualStakingMaxCapLimitsRequest struct{}

func (m *QueryVirtualStakingMaxCapLimitsRequest) Reset() {
	*m = QueryVirtualStakingMaxCapLimitsRequest{}
}
func (m *QueryVirtualStakingMaxCapLimitsRequest) String() string { return proto.CompactTextString(m) }
func (*QueryVirtualStakingMaxCapLimitsRequest) ProtoMessage()    {}
func (*QueryVirtualStakingMaxCapLimitsRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_50c89ba006eed4fb, []int{2}
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryVirtualStakingMaxCapLimitsRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryVirtualStakingMaxCapLimitsRequest.Merge(m, src)
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) XXX_Size() int {
	return m.Size()
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryVirtualStakingMaxCapLimitsRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryVirtualStakingMaxCapLimitsRequest proto.InternalMessageInfo

// QueryVirtualStakingMaxCapLimitsResponse is the response type for the
// Query/VirtualStakingMaxCapLimits RPC method
type QueryVirtualStakingMaxCapLimitsResponse struct {
	MaxCapInfos []VirtualStakingMaxCapInfo `protobuf:"bytes,1,rep,name=max_cap_infos,json=maxCapInfos,proto3" json:"max_cap_infos"`
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) Reset() {
	*m = QueryVirtualStakingMaxCapLimitsResponse{}
}
func (m *QueryVirtualStakingMaxCapLimitsResponse) String() string { return proto.CompactTextString(m) }
func (*QueryVirtualStakingMaxCapLimitsResponse) ProtoMessage()    {}
func (*QueryVirtualStakingMaxCapLimitsResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_50c89ba006eed4fb, []int{3}
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryVirtualStakingMaxCapLimitsResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryVirtualStakingMaxCapLimitsResponse.Merge(m, src)
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) XXX_Size() int {
	return m.Size()
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryVirtualStakingMaxCapLimitsResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryVirtualStakingMaxCapLimitsResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*QueryVirtualStakingMaxCapLimitRequest)(nil), "osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitRequest")
	proto.RegisterType((*QueryVirtualStakingMaxCapLimitResponse)(nil), "osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitResponse")
	proto.RegisterType((*QueryVirtualStakingMaxCapLimitsRequest)(nil), "osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitsRequest")
	proto.RegisterType((*QueryVirtualStakingMaxCapLimitsResponse)(nil), "osmosis.meshsecurity.v1beta1.QueryVirtualStakingMaxCapLimitsResponse")
}

func init() {
	proto.RegisterFile("osmosis/meshsecurity/v1beta1/query.proto", fileDescriptor_50c89ba006eed4fb)
}

var fileDescriptor_50c89ba006eed4fb = []byte{
	// 503 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xa4, 0x53, 0x4d, 0x6b, 0x13, 0x41,
	0x18, 0xde, 0x69, 0xaa, 0x92, 0x09, 0x1e, 0x1c, 0x3c, 0xa4, 0x4b, 0xd9, 0xca, 0x82, 0x1a, 0xc4,
	0xcc, 0x90, 0xf8, 0x05, 0x42, 0x05, 0x13, 0x3d, 0x08, 0x7a, 0x30, 0x82, 0x07, 0x2f, 0x75, 0x76,
	0x77, 0xba, 0x1d, 0xba, 0x3b, 0xb3, 0xdd, 0x99, 0x95, 0x04, 0xf1, 0xe2, 0x2f, 0x28, 0xf8, 0x07,
	0x7a, 0xec, 0xd1, 0x9f, 0x91, 0x63, 0xc5, 0x8b, 0x27, 0x3f, 0x12, 0x45, 0x7f, 0x86, 0xec, 0xee,
	0x6c, 0xb5, 0x60, 0x93, 0x48, 0x2e, 0x61, 0xf2, 0xce, 0xfb, 0x3c, 0xef, 0xf3, 0x3c, 0xf3, 0x2e,
	0x6c, 0x49, 0x15, 0x4b, 0xc5, 0x15, 0x89, 0x99, 0xda, 0x51, 0xcc, 0xcf, 0x52, 0xae, 0x47, 0xe4,
	0x55, 0xc7, 0x63, 0x9a, 0x76, 0xc8, 0x5e, 0xc6, 0xd2, 0x11, 0x4e, 0x52, 0xa9, 0x25, 0x5a, 0x37,
	0x9d, 0xf8, 0xef, 0x4e, 0x6c, 0x3a, 0x6d, 0xc7, 0x2f, 0xae, 0x89, 0x47, 0x15, 0x3b, 0x86, 0xfb,
	0x92, 0x8b, 0x12, 0x6d, 0x93, 0x99, 0x73, 0x4e, 0x50, 0x96, 0x80, 0x8b, 0xa1, 0x0c, 0x65, 0x71,
	0x24, 0xf9, 0xc9, 0x54, 0xd7, 0x43, 0x29, 0xc3, 0x88, 0x11, 0x9a, 0x70, 0x42, 0x85, 0x90, 0x9a,
	0x6a, 0x2e, 0x85, 0x32, 0xb7, 0x17, 0x68, 0xcc, 0x85, 0x24, 0xc5, 0x6f, 0x59, 0x72, 0xef, 0xc3,
	0xcb, 0x4f, 0x73, 0x13, 0xcf, 0x79, 0xaa, 0x33, 0x1a, 0x3d, 0xd3, 0x74, 0x97, 0x8b, 0xf0, 0x09,
	0x1d, 0xf6, 0x69, 0xf2, 0x98, 0xc7, 0x5c, 0x0f, 0xd8, 0x5e, 0xc6, 0x94, 0x46, 0x4d, 0x78, 0x8e,
	0x06, 0x41, 0xca, 0x94, 0x6a, 0x82, 0x4b, 0xa0, 0x55, 0x1f, 0x54, 0x7f, 0xdd, 0x03, 0x00, 0xaf,
	0xcc, 0xe3, 0x50, 0x89, 0x14, 0x8a, 0xa1, 0x4d, 0x58, 0x0f, 0x58, 0xc4, 0x42, 0xaa, 0x59, 0x50,
	0xd0, 0x34, 0xba, 0x6b, 0xb8, 0x4c, 0x06, 0xe7, 0xc9, 0x54, 0x71, 0xe1, 0xbe, 0xe4, 0xa2, 0xb7,
	0x3a, 0xfe, 0xbc, 0x61, 0x0d, 0xfe, 0x20, 0x50, 0x07, 0xd6, 0x7c, 0x9a, 0x34, 0x57, 0x16, 0x03,
	0xe6, 0xbd, 0x77, 0x57, 0x7f, 0x1d, 0x6c, 0x00, 0xb7, 0x35, 0x4f, 0xa1, 0x32, 0x36, 0xdd, 0x7d,
	0x00, 0xaf, 0xce, 0x6d, 0x35, 0x6e, 0x18, 0x3c, 0x1f, 0xd3, 0xe1, 0x96, 0x4f, 0x93, 0x2d, 0x2e,
	0xb6, 0x65, 0x1e, 0x4c, 0xad, 0xd5, 0xe8, 0xde, 0xc6, 0xb3, 0x36, 0x01, 0xff, 0x8b, 0xf8, 0x91,
	0xd8, 0x96, 0xbd, 0x7a, 0xae, 0xfa, 0xf0, 0xe7, 0xfb, 0x6b, 0x60, 0xd0, 0x88, 0x8f, 0xcb, 0xaa,
	0xfb, 0xa1, 0x06, 0xcf, 0x14, 0x92, 0xd0, 0x0f, 0x00, 0xd7, 0x4e, 0xd5, 0x85, 0xfa, 0xb3, 0xe7,
	0x2e, 0xf4, 0xcc, 0xf6, 0x83, 0xe5, 0x48, 0xca, 0x64, 0xdc, 0xcd, 0xb7, 0x1f, 0xbf, 0xbf, 0x5b,
	0xb9, 0x83, 0x6e, 0xcd, 0x59, 0x6b, 0x93, 0x5e, 0x94, 0x83, 0xc9, 0x6b, 0xb3, 0x50, 0x6f, 0xd0,
	0x17, 0x00, 0xed, 0xd3, 0xf3, 0x47, 0x4b, 0x69, 0xac, 0x5e, 0xda, 0x7e, 0xb8, 0x24, 0x8b, 0xb1,
	0x7a, 0xb3, 0xb0, 0x8a, 0xd1, 0xf5, 0xff, 0xb0, 0xaa, 0x7a, 0x2f, 0xc7, 0xdf, 0x1c, 0xeb, 0x70,
	0xe2, 0x58, 0xe3, 0x89, 0x03, 0x8e, 0x26, 0x0e, 0xf8, 0x3a, 0x71, 0xc0, 0xfe, 0xd4, 0xb1, 0x8e,
	0xa6, 0x8e, 0xf5, 0x69, 0xea, 0x58, 0x2f, 0xee, 0x85, 0x5c, 0xef, 0x64, 0x1e, 0xf6, 0x65, 0x5c,
	0x31, 0xb7, 0x23, 0xea, 0x95, 0xf4, 0xed, 0x8a, 0xbf, 0xad, 0x82, 0x5d, 0x32, 0x3c, 0x39, 0x52,
	0x8f, 0x12, 0xa6, 0xbc, 0xb3, 0xc5, 0xf7, 0x7d, 0xe3, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x3f,
	0x3e, 0x93, 0x23, 0xc1, 0x04, 0x00, 0x00,
}

func (this *QueryVirtualStakingMaxCapLimitResponse) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*QueryVirtualStakingMaxCapLimitResponse)
	if !ok {
		that2, ok := that.(QueryVirtualStakingMaxCapLimitResponse)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !this.Delegated.Equal(&that1.Delegated) {
		return false
	}
	if !this.Cap.Equal(&that1.Cap) {
		return false
	}
	return true
}

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ context.Context
	_ grpc.ClientConn
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// VirtualStakingMaxCapLimit gets max cap limit for the given contract
	VirtualStakingMaxCapLimit(ctx context.Context, in *QueryVirtualStakingMaxCapLimitRequest, opts ...grpc.CallOption) (*QueryVirtualStakingMaxCapLimitResponse, error)
	// VirtualStakingMaxCapLimits gets max cap limits
	VirtualStakingMaxCapLimits(ctx context.Context, in *QueryVirtualStakingMaxCapLimitsRequest, opts ...grpc.CallOption) (*QueryVirtualStakingMaxCapLimitsResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) VirtualStakingMaxCapLimit(ctx context.Context, in *QueryVirtualStakingMaxCapLimitRequest, opts ...grpc.CallOption) (*QueryVirtualStakingMaxCapLimitResponse, error) {
	out := new(QueryVirtualStakingMaxCapLimitResponse)
	err := c.cc.Invoke(ctx, "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCapLimit", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) VirtualStakingMaxCapLimits(ctx context.Context, in *QueryVirtualStakingMaxCapLimitsRequest, opts ...grpc.CallOption) (*QueryVirtualStakingMaxCapLimitsResponse, error) {
	out := new(QueryVirtualStakingMaxCapLimitsResponse)
	err := c.cc.Invoke(ctx, "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCapLimits", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// VirtualStakingMaxCapLimit gets max cap limit for the given contract
	VirtualStakingMaxCapLimit(context.Context, *QueryVirtualStakingMaxCapLimitRequest) (*QueryVirtualStakingMaxCapLimitResponse, error)
	// VirtualStakingMaxCapLimits gets max cap limits
	VirtualStakingMaxCapLimits(context.Context, *QueryVirtualStakingMaxCapLimitsRequest) (*QueryVirtualStakingMaxCapLimitsResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct{}

func (*UnimplementedQueryServer) VirtualStakingMaxCapLimit(ctx context.Context, req *QueryVirtualStakingMaxCapLimitRequest) (*QueryVirtualStakingMaxCapLimitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VirtualStakingMaxCapLimit not implemented")
}

func (*UnimplementedQueryServer) VirtualStakingMaxCapLimits(ctx context.Context, req *QueryVirtualStakingMaxCapLimitsRequest) (*QueryVirtualStakingMaxCapLimitsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method VirtualStakingMaxCapLimits not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_VirtualStakingMaxCapLimit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryVirtualStakingMaxCapLimitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).VirtualStakingMaxCapLimit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCapLimit",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).VirtualStakingMaxCapLimit(ctx, req.(*QueryVirtualStakingMaxCapLimitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_VirtualStakingMaxCapLimits_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryVirtualStakingMaxCapLimitsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).VirtualStakingMaxCapLimits(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.meshsecurity.v1beta1.Query/VirtualStakingMaxCapLimits",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).VirtualStakingMaxCapLimits(ctx, req.(*QueryVirtualStakingMaxCapLimitsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.meshsecurity.v1beta1.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "VirtualStakingMaxCapLimit",
			Handler:    _Query_VirtualStakingMaxCapLimit_Handler,
		},
		{
			MethodName: "VirtualStakingMaxCapLimits",
			Handler:    _Query_VirtualStakingMaxCapLimits_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/meshsecurity/v1beta1/query.proto",
}

func (m *QueryVirtualStakingMaxCapLimitRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryVirtualStakingMaxCapLimitRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryVirtualStakingMaxCapLimitRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Address) > 0 {
		i -= len(m.Address)
		copy(dAtA[i:], m.Address)
		i = encodeVarintQuery(dAtA, i, uint64(len(m.Address)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryVirtualStakingMaxCapLimitResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryVirtualStakingMaxCapLimitResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryVirtualStakingMaxCapLimitResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Cap.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size, err := m.Delegated.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.MaxCapInfos) > 0 {
		for iNdEx := len(m.MaxCapInfos) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.MaxCapInfos[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}

func (m *QueryVirtualStakingMaxCapLimitRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Address)
	if l > 0 {
		n += 1 + l + sovQuery(uint64(l))
	}
	return n
}

func (m *QueryVirtualStakingMaxCapLimitResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Delegated.Size()
	n += 1 + l + sovQuery(uint64(l))
	l = m.Cap.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.MaxCapInfos) > 0 {
		for _, e := range m.MaxCapInfos {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}

func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}

func (m *QueryVirtualStakingMaxCapLimitRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryVirtualStakingMaxCapLimitRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryVirtualStakingMaxCapLimitRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Address", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Address = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (m *QueryVirtualStakingMaxCapLimitResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryVirtualStakingMaxCapLimitResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryVirtualStakingMaxCapLimitResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Delegated", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Delegated.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cap", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Cap.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (m *QueryVirtualStakingMaxCapLimitsRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryVirtualStakingMaxCapLimitsRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryVirtualStakingMaxCapLimitsRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func (m *QueryVirtualStakingMaxCapLimitsResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryVirtualStakingMaxCapLimitsResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryVirtualStakingMaxCapLimitsResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxCapInfos", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.MaxCapInfos = append(m.MaxCapInfos, VirtualStakingMaxCapInfo{})
			if err := m.MaxCapInfos[len(m.MaxCapInfos)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
