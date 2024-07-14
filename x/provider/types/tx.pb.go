// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/provider/v1beta1/tx.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	github_com_cosmos_cosmos_sdk_types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/msgservice"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type MsgSetConsumerCommissionRate struct {
	ProviderAddr string                                 `protobuf:"bytes,1,opt,name=provider_addr,json=providerAddr,proto3" json:"provider_addr,omitempty" yaml:"address"`
	ChainId      string                                 `protobuf:"bytes,2,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	Rate         github_com_cosmos_cosmos_sdk_types.Dec `protobuf:"bytes,3,opt,name=rate,proto3,customtype=github.com/cosmos/cosmos-sdk/types.Dec" json:"rate"`
}

func (m *MsgSetConsumerCommissionRate) Reset()         { *m = MsgSetConsumerCommissionRate{} }
func (m *MsgSetConsumerCommissionRate) String() string { return proto.CompactTextString(m) }
func (*MsgSetConsumerCommissionRate) ProtoMessage()    {}
func (*MsgSetConsumerCommissionRate) Descriptor() ([]byte, []int) {
	return fileDescriptor_600cc2e6a7e19cb9, []int{0}
}
func (m *MsgSetConsumerCommissionRate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSetConsumerCommissionRate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSetConsumerCommissionRate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSetConsumerCommissionRate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSetConsumerCommissionRate.Merge(m, src)
}
func (m *MsgSetConsumerCommissionRate) XXX_Size() int {
	return m.Size()
}
func (m *MsgSetConsumerCommissionRate) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSetConsumerCommissionRate.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSetConsumerCommissionRate proto.InternalMessageInfo

type MsgSetConsumerCommissionRateResponse struct {
}

func (m *MsgSetConsumerCommissionRateResponse) Reset()         { *m = MsgSetConsumerCommissionRateResponse{} }
func (m *MsgSetConsumerCommissionRateResponse) String() string { return proto.CompactTextString(m) }
func (*MsgSetConsumerCommissionRateResponse) ProtoMessage()    {}
func (*MsgSetConsumerCommissionRateResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_600cc2e6a7e19cb9, []int{1}
}
func (m *MsgSetConsumerCommissionRateResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgSetConsumerCommissionRateResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgSetConsumerCommissionRateResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgSetConsumerCommissionRateResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgSetConsumerCommissionRateResponse.Merge(m, src)
}
func (m *MsgSetConsumerCommissionRateResponse) XXX_Size() int {
	return m.Size()
}
func (m *MsgSetConsumerCommissionRateResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgSetConsumerCommissionRateResponse.DiscardUnknown(m)
}

var xxx_messageInfo_MsgSetConsumerCommissionRateResponse proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgSetConsumerCommissionRate)(nil), "osmosis.provider.v1beta1.MsgSetConsumerCommissionRate")
	proto.RegisterType((*MsgSetConsumerCommissionRateResponse)(nil), "osmosis.provider.v1beta1.MsgSetConsumerCommissionRateResponse")
}

func init() { proto.RegisterFile("osmosis/provider/v1beta1/tx.proto", fileDescriptor_600cc2e6a7e19cb9) }

var fileDescriptor_600cc2e6a7e19cb9 = []byte{
	// 404 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0xcc, 0x2f, 0xce, 0xcd,
	0x2f, 0xce, 0x2c, 0xd6, 0x2f, 0x28, 0xca, 0x2f, 0xcb, 0x4c, 0x49, 0x2d, 0xd2, 0x2f, 0x33, 0x4c,
	0x4a, 0x2d, 0x49, 0x34, 0xd4, 0x2f, 0xa9, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x92, 0x80,
	0x2a, 0xd1, 0x83, 0x29, 0xd1, 0x83, 0x2a, 0x91, 0x92, 0x4c, 0x06, 0x4b, 0xc5, 0x83, 0xd5, 0xe9,
	0x43, 0x38, 0x10, 0x4d, 0x52, 0xe2, 0x10, 0x9e, 0x7e, 0x6e, 0x71, 0xba, 0x7e, 0x99, 0x21, 0x88,
	0x82, 0x4a, 0x88, 0xa4, 0xe7, 0xa7, 0xe7, 0x43, 0x34, 0x80, 0x58, 0x50, 0x51, 0xc1, 0xc4, 0xdc,
	0xcc, 0xbc, 0x7c, 0x7d, 0x30, 0x09, 0x11, 0x52, 0x3a, 0xc7, 0xc8, 0x25, 0xe3, 0x5b, 0x9c, 0x1e,
	0x9c, 0x5a, 0xe2, 0x9c, 0x9f, 0x57, 0x5c, 0x9a, 0x9b, 0x5a, 0xe4, 0x9c, 0x9f, 0x9b, 0x9b, 0x59,
	0x5c, 0x9c, 0x99, 0x9f, 0x17, 0x94, 0x58, 0x92, 0x2a, 0x64, 0xce, 0xc5, 0x0b, 0x73, 0x51, 0x7c,
	0x62, 0x4a, 0x4a, 0x91, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0xa7, 0x93, 0xd0, 0xa7, 0x7b, 0xf2, 0x7c,
	0x95, 0x89, 0xb9, 0x39, 0x56, 0x4a, 0x20, 0xd1, 0xd4, 0xe2, 0x62, 0xa5, 0x20, 0x1e, 0x98, 0x42,
	0xc7, 0x94, 0x94, 0x22, 0x21, 0x49, 0x2e, 0x8e, 0xe4, 0x8c, 0xc4, 0xcc, 0xbc, 0xf8, 0xcc, 0x14,
	0x09, 0x26, 0x90, 0x9e, 0x20, 0x76, 0x30, 0xdf, 0x33, 0x45, 0x28, 0x80, 0x8b, 0xa5, 0x28, 0xb1,
	0x24, 0x55, 0x82, 0x19, 0x6c, 0x94, 0xcd, 0x89, 0x7b, 0xf2, 0x0c, 0xb7, 0xee, 0xc9, 0xab, 0xa5,
	0x67, 0x96, 0x64, 0x94, 0x26, 0xe9, 0x25, 0xe7, 0xe7, 0x42, 0x7d, 0x09, 0xa5, 0x74, 0x8b, 0x53,
	0xb2, 0xf5, 0x4b, 0x2a, 0x0b, 0x52, 0x8b, 0xf5, 0x5c, 0x52, 0x93, 0x2f, 0x6d, 0xd1, 0xe5, 0x82,
	0x06, 0x82, 0x4b, 0x6a, 0x72, 0x10, 0xd8, 0x24, 0x2b, 0x8e, 0x8e, 0x05, 0xf2, 0x0c, 0x2f, 0x16,
	0xc8, 0x33, 0x28, 0xa9, 0x71, 0xa9, 0xe0, 0xf3, 0x4f, 0x50, 0x6a, 0x71, 0x41, 0x7e, 0x5e, 0x71,
	0xaa, 0xd1, 0x6c, 0x46, 0x2e, 0x66, 0xdf, 0xe2, 0x74, 0xa1, 0xc9, 0x8c, 0x5c, 0x92, 0xb8, 0x7d,
	0x6f, 0xa6, 0x87, 0x2b, 0x5a, 0xf4, 0xf0, 0xd9, 0x22, 0x65, 0x47, 0x9e, 0x3e, 0x98, 0xeb, 0x9c,
	0x22, 0x4e, 0x3c, 0x94, 0x63, 0x38, 0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23, 0x39, 0xc6, 0x07, 0x8f,
	0xe4, 0x18, 0x27, 0x3c, 0x96, 0x63, 0xb8, 0xf0, 0x58, 0x8e, 0xe1, 0xc6, 0x63, 0x39, 0x86, 0x28,
	0x2b, 0xa4, 0x90, 0x82, 0xda, 0xa3, 0x9b, 0x93, 0x98, 0x54, 0xac, 0x9f, 0x9b, 0x5a, 0x9c, 0xa1,
	0x5b, 0x9c, 0x9a, 0x5c, 0x5a, 0x94, 0x59, 0x52, 0x09, 0x0e, 0xb6, 0x0a, 0x44, 0x92, 0x03, 0x87,
	0x60, 0x12, 0x1b, 0x38, 0xde, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0xaf, 0x67, 0x1e, 0x99,
	0x93, 0x02, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MsgClient is the client API for Msg service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MsgClient interface {
	SetConsumerCommissionRate(ctx context.Context, in *MsgSetConsumerCommissionRate, opts ...grpc.CallOption) (*MsgSetConsumerCommissionRateResponse, error)
}

type msgClient struct {
	cc grpc1.ClientConn
}

func NewMsgClient(cc grpc1.ClientConn) MsgClient {
	return &msgClient{cc}
}

func (c *msgClient) SetConsumerCommissionRate(ctx context.Context, in *MsgSetConsumerCommissionRate, opts ...grpc.CallOption) (*MsgSetConsumerCommissionRateResponse, error) {
	out := new(MsgSetConsumerCommissionRateResponse)
	err := c.cc.Invoke(ctx, "/osmosis.provider.v1beta1.Msg/SetConsumerCommissionRate", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MsgServer is the server API for Msg service.
type MsgServer interface {
	SetConsumerCommissionRate(context.Context, *MsgSetConsumerCommissionRate) (*MsgSetConsumerCommissionRateResponse, error)
}

// UnimplementedMsgServer can be embedded to have forward compatible implementations.
type UnimplementedMsgServer struct {
}

func (*UnimplementedMsgServer) SetConsumerCommissionRate(ctx context.Context, req *MsgSetConsumerCommissionRate) (*MsgSetConsumerCommissionRateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetConsumerCommissionRate not implemented")
}

func RegisterMsgServer(s grpc1.Server, srv MsgServer) {
	s.RegisterService(&_Msg_serviceDesc, srv)
}

func _Msg_SetConsumerCommissionRate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MsgSetConsumerCommissionRate)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MsgServer).SetConsumerCommissionRate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/osmosis.provider.v1beta1.Msg/SetConsumerCommissionRate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MsgServer).SetConsumerCommissionRate(ctx, req.(*MsgSetConsumerCommissionRate))
	}
	return interceptor(ctx, in, info, handler)
}

var _Msg_serviceDesc = grpc.ServiceDesc{
	ServiceName: "osmosis.provider.v1beta1.Msg",
	HandlerType: (*MsgServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetConsumerCommissionRate",
			Handler:    _Msg_SetConsumerCommissionRate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "osmosis/provider/v1beta1/tx.proto",
}

func (m *MsgSetConsumerCommissionRate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSetConsumerCommissionRate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSetConsumerCommissionRate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.Rate.Size()
		i -= size
		if _, err := m.Rate.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintTx(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.ChainId) > 0 {
		i -= len(m.ChainId)
		copy(dAtA[i:], m.ChainId)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ChainId)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.ProviderAddr) > 0 {
		i -= len(m.ProviderAddr)
		copy(dAtA[i:], m.ProviderAddr)
		i = encodeVarintTx(dAtA, i, uint64(len(m.ProviderAddr)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *MsgSetConsumerCommissionRateResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgSetConsumerCommissionRateResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgSetConsumerCommissionRateResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func encodeVarintTx(dAtA []byte, offset int, v uint64) int {
	offset -= sovTx(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgSetConsumerCommissionRate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.ProviderAddr)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = len(m.ChainId)
	if l > 0 {
		n += 1 + l + sovTx(uint64(l))
	}
	l = m.Rate.Size()
	n += 1 + l + sovTx(uint64(l))
	return n
}

func (m *MsgSetConsumerCommissionRateResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func sovTx(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozTx(x uint64) (n int) {
	return sovTx(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgSetConsumerCommissionRate) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgSetConsumerCommissionRate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSetConsumerCommissionRate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ProviderAddr", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ProviderAddr = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainId", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ChainId = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Rate", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowTx
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
				return ErrInvalidLengthTx
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthTx
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Rate.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func (m *MsgSetConsumerCommissionRateResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowTx
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
			return fmt.Errorf("proto: MsgSetConsumerCommissionRateResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgSetConsumerCommissionRateResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipTx(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthTx
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
func skipTx(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
					return 0, ErrIntOverflowTx
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
				return 0, ErrInvalidLengthTx
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupTx
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthTx
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthTx        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowTx          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupTx = fmt.Errorf("proto: unexpected end of group")
)
