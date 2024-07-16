// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: osmosis/meshsecurity/v1beta1/meshsecurity.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/cosmos-sdk/types/tx/amino"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

// VirtualStakingMaxCapInfo stores info about
// virtual staking max cap
type VirtualStakingMaxCapInfo struct {
	// Contract is the address of the contract
	Contract string `protobuf:"bytes,1,opt,name=contract,proto3" json:"contract,omitempty"`
	// Delegated is the total amount currently delegated
	Delegated types.Coin `protobuf:"bytes,2,opt,name=delegated,proto3" json:"delegated"`
	// Cap is the current max cap limit
	Cap types.Coin `protobuf:"bytes,3,opt,name=cap,proto3" json:"cap"`
}

func (m *VirtualStakingMaxCapInfo) Reset()         { *m = VirtualStakingMaxCapInfo{} }
func (m *VirtualStakingMaxCapInfo) String() string { return proto.CompactTextString(m) }
func (*VirtualStakingMaxCapInfo) ProtoMessage()    {}
func (*VirtualStakingMaxCapInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_53771980e3e4256c, []int{0}
}
func (m *VirtualStakingMaxCapInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *VirtualStakingMaxCapInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_VirtualStakingMaxCapInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *VirtualStakingMaxCapInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VirtualStakingMaxCapInfo.Merge(m, src)
}
func (m *VirtualStakingMaxCapInfo) XXX_Size() int {
	return m.Size()
}
func (m *VirtualStakingMaxCapInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_VirtualStakingMaxCapInfo.DiscardUnknown(m)
}

var xxx_messageInfo_VirtualStakingMaxCapInfo proto.InternalMessageInfo

// Params defines the parameters for the x/meshsecurity module.
type Params struct {
	// TotalContractsMaxCap is the maximum that the sum of all contract max caps
	// must not exceed
	TotalContractsMaxCap types.Coin `protobuf:"bytes,1,opt,name=total_contracts_max_cap,json=totalContractsMaxCap,proto3" json:"total_contracts_max_cap"`
	// Epoch length is the number of blocks that defines an epoch
	EpochLength uint32 `protobuf:"varint,2,opt,name=epoch_length,json=epochLength,proto3" json:"epoch_length,omitempty"`
	// MaxGasEndBlocker defines the maximum gas that can be spent in a contract
	// sudo callback
	MaxGasEndBlocker uint32 `protobuf:"varint,3,opt,name=max_gas_end_blocker,json=maxGasEndBlocker,proto3" json:"max_gas_end_blocker,omitempty"`
	InfractionTime   uint64 `protobuf:"varint,4,opt,name=infraction_time,json=infractionTime,proto3" json:"infraction_time,omitempty"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_53771980e3e4256c, []int{1}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func init() {
	proto.RegisterType((*VirtualStakingMaxCapInfo)(nil), "osmosis.meshsecurity.v1beta1.VirtualStakingMaxCapInfo")
	proto.RegisterType((*Params)(nil), "osmosis.meshsecurity.v1beta1.Params")
}

func init() {
	proto.RegisterFile("osmosis/meshsecurity/v1beta1/meshsecurity.proto", fileDescriptor_53771980e3e4256c)
}

var fileDescriptor_53771980e3e4256c = []byte{
	// 444 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0x3f, 0x6f, 0xd3, 0x40,
	0x1c, 0xf5, 0xd1, 0xa8, 0xa2, 0x57, 0xca, 0x9f, 0x6b, 0x25, 0x4c, 0x54, 0x5d, 0x43, 0x17, 0x22,
	0xa4, 0xd8, 0x0a, 0x6c, 0x95, 0x60, 0x48, 0x84, 0x10, 0x12, 0x48, 0xc8, 0xa0, 0x0e, 0x2c, 0xe6,
	0xe7, 0xf3, 0xd5, 0x39, 0xc5, 0x77, 0x67, 0xf9, 0x2e, 0x28, 0xfd, 0x0a, 0x4c, 0x7c, 0x04, 0x46,
	0x26, 0xc4, 0xc7, 0xc8, 0xd8, 0x91, 0x09, 0x41, 0x32, 0xc0, 0x57, 0x60, 0xab, 0x7c, 0x8e, 0x5b,
	0x79, 0xcb, 0x62, 0xfd, 0xfc, 0x7e, 0xf7, 0xde, 0xef, 0x3d, 0xe9, 0xe1, 0x50, 0x1b, 0xa9, 0x8d,
	0x30, 0xa1, 0xe4, 0x66, 0x62, 0x38, 0x9b, 0x95, 0xc2, 0x9e, 0x87, 0x9f, 0x86, 0x09, 0xb7, 0x30,
	0x6c, 0x81, 0x41, 0x51, 0x6a, 0xab, 0xc9, 0xe1, 0x9a, 0x10, 0xb4, 0x76, 0x6b, 0x42, 0x97, 0x32,
	0xb7, 0x0e, 0x13, 0x30, 0xfc, 0x4a, 0x85, 0x69, 0xa1, 0x6a, 0x76, 0xf7, 0x20, 0xd3, 0x99, 0x76,
	0x63, 0x58, 0x4d, 0x6b, 0xf4, 0x1e, 0x48, 0xa1, 0x74, 0xe8, 0xbe, 0x35, 0x74, 0xfc, 0x1d, 0x61,
	0xff, 0x54, 0x94, 0x76, 0x06, 0xf9, 0x3b, 0x0b, 0x53, 0xa1, 0xb2, 0x37, 0x30, 0x1f, 0x43, 0xf1,
	0x4a, 0x9d, 0x69, 0xd2, 0xc5, 0x37, 0x99, 0x56, 0xb6, 0x04, 0x66, 0x7d, 0xd4, 0x43, 0xfd, 0x9d,
	0xe8, 0xea, 0x9f, 0x3c, 0xc3, 0x3b, 0x29, 0xcf, 0x79, 0x06, 0x96, 0xa7, 0xfe, 0x8d, 0x1e, 0xea,
	0xef, 0x3e, 0x79, 0x10, 0xd4, 0xae, 0x82, 0xca, 0x55, 0x63, 0x35, 0x18, 0x6b, 0xa1, 0x46, 0x9d,
	0xc5, 0xaf, 0x23, 0x2f, 0xba, 0x66, 0x90, 0x21, 0xde, 0x62, 0x50, 0xf8, 0x5b, 0x9b, 0x11, 0xab,
	0xb7, 0x27, 0x9d, 0x7f, 0x5f, 0x8f, 0xd0, 0xf1, 0x7f, 0x84, 0xb7, 0xdf, 0x42, 0x09, 0xd2, 0x90,
	0x53, 0x7c, 0xdf, 0x6a, 0x0b, 0x79, 0xdc, 0x98, 0x32, 0xb1, 0x84, 0x79, 0x5c, 0xe9, 0xa2, 0xcd,
	0x74, 0x0f, 0x1c, 0x7f, 0xdc, 0xd0, 0xeb, 0xe8, 0xe4, 0x21, 0xbe, 0xc5, 0x0b, 0xcd, 0x26, 0x71,
	0xce, 0x55, 0x66, 0x27, 0x2e, 0xdd, 0x5e, 0xb4, 0xeb, 0xb0, 0xd7, 0x0e, 0x22, 0x03, 0xbc, 0x5f,
	0x9d, 0xca, 0xc0, 0xc4, 0x5c, 0xa5, 0x71, 0x92, 0x6b, 0x36, 0xe5, 0xa5, 0x8b, 0xb3, 0x17, 0xdd,
	0x95, 0x30, 0x7f, 0x09, 0xe6, 0x85, 0x4a, 0x47, 0x35, 0x4e, 0x1e, 0xe1, 0x3b, 0x42, 0x9d, 0x55,
	0x37, 0x84, 0x56, 0xb1, 0x15, 0x92, 0xfb, 0x9d, 0x1e, 0xea, 0x77, 0xa2, 0xdb, 0xd7, 0xf0, 0x7b,
	0x21, 0xf9, 0xc9, 0x61, 0x95, 0xf1, 0xf3, 0xdf, 0x1f, 0x8f, 0xf7, 0x5b, 0x3d, 0xa9, 0x03, 0x8f,
	0x3e, 0x2e, 0xfe, 0x50, 0xef, 0xdb, 0x92, 0x7a, 0x8b, 0x25, 0x45, 0x17, 0x4b, 0x8a, 0x7e, 0x2f,
	0x29, 0xfa, 0xb2, 0xa2, 0xde, 0xc5, 0x8a, 0x7a, 0x3f, 0x57, 0xd4, 0xfb, 0xf0, 0x3c, 0x13, 0x76,
	0x32, 0x4b, 0x02, 0xa6, 0x65, 0xd3, 0xb8, 0x41, 0x0e, 0x49, 0x5d, 0xbb, 0x41, 0xa3, 0x37, 0x30,
	0xe9, 0x34, 0x9c, 0xb7, 0xab, 0x68, 0xcf, 0x0b, 0x6e, 0x92, 0x6d, 0xd7, 0x8a, 0xa7, 0x97, 0x01,
	0x00, 0x00, 0xff, 0xff, 0x95, 0xad, 0x29, 0xf1, 0xaf, 0x02, 0x00, 0x00,
}

func (this *VirtualStakingMaxCapInfo) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*VirtualStakingMaxCapInfo)
	if !ok {
		that2, ok := that.(VirtualStakingMaxCapInfo)
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
	if this.Contract != that1.Contract {
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
func (this *Params) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*Params)
	if !ok {
		that2, ok := that.(Params)
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
	if !this.TotalContractsMaxCap.Equal(&that1.TotalContractsMaxCap) {
		return false
	}
	if this.EpochLength != that1.EpochLength {
		return false
	}
	if this.MaxGasEndBlocker != that1.MaxGasEndBlocker {
		return false
	}
	if this.InfractionTime != that1.InfractionTime {
		return false
	}
	return true
}
func (m *VirtualStakingMaxCapInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *VirtualStakingMaxCapInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *VirtualStakingMaxCapInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
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
		i = encodeVarintMeshsecurity(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size, err := m.Delegated.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintMeshsecurity(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Contract) > 0 {
		i -= len(m.Contract)
		copy(dAtA[i:], m.Contract)
		i = encodeVarintMeshsecurity(dAtA, i, uint64(len(m.Contract)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.InfractionTime != 0 {
		i = encodeVarintMeshsecurity(dAtA, i, uint64(m.InfractionTime))
		i--
		dAtA[i] = 0x20
	}
	if m.MaxGasEndBlocker != 0 {
		i = encodeVarintMeshsecurity(dAtA, i, uint64(m.MaxGasEndBlocker))
		i--
		dAtA[i] = 0x18
	}
	if m.EpochLength != 0 {
		i = encodeVarintMeshsecurity(dAtA, i, uint64(m.EpochLength))
		i--
		dAtA[i] = 0x10
	}
	{
		size, err := m.TotalContractsMaxCap.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintMeshsecurity(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintMeshsecurity(dAtA []byte, offset int, v uint64) int {
	offset -= sovMeshsecurity(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *VirtualStakingMaxCapInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Contract)
	if l > 0 {
		n += 1 + l + sovMeshsecurity(uint64(l))
	}
	l = m.Delegated.Size()
	n += 1 + l + sovMeshsecurity(uint64(l))
	l = m.Cap.Size()
	n += 1 + l + sovMeshsecurity(uint64(l))
	return n
}

func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.TotalContractsMaxCap.Size()
	n += 1 + l + sovMeshsecurity(uint64(l))
	if m.EpochLength != 0 {
		n += 1 + sovMeshsecurity(uint64(m.EpochLength))
	}
	if m.MaxGasEndBlocker != 0 {
		n += 1 + sovMeshsecurity(uint64(m.MaxGasEndBlocker))
	}
	if m.InfractionTime != 0 {
		n += 1 + sovMeshsecurity(uint64(m.InfractionTime))
	}
	return n
}

func sovMeshsecurity(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMeshsecurity(x uint64) (n int) {
	return sovMeshsecurity(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *VirtualStakingMaxCapInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMeshsecurity
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
			return fmt.Errorf("proto: VirtualStakingMaxCapInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: VirtualStakingMaxCapInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Contract", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshsecurity
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
				return ErrInvalidLengthMeshsecurity
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMeshsecurity
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Contract = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Delegated", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshsecurity
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
				return ErrInvalidLengthMeshsecurity
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeshsecurity
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Delegated.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Cap", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshsecurity
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
				return ErrInvalidLengthMeshsecurity
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeshsecurity
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
			skippy, err := skipMeshsecurity(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMeshsecurity
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
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMeshsecurity
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TotalContractsMaxCap", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshsecurity
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
				return ErrInvalidLengthMeshsecurity
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMeshsecurity
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TotalContractsMaxCap.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field EpochLength", wireType)
			}
			m.EpochLength = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshsecurity
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.EpochLength |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field MaxGasEndBlocker", wireType)
			}
			m.MaxGasEndBlocker = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshsecurity
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.MaxGasEndBlocker |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field InfractionTime", wireType)
			}
			m.InfractionTime = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMeshsecurity
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.InfractionTime |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		default:
			iNdEx = preIndex
			skippy, err := skipMeshsecurity(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthMeshsecurity
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
func skipMeshsecurity(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMeshsecurity
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
					return 0, ErrIntOverflowMeshsecurity
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
					return 0, ErrIntOverflowMeshsecurity
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
				return 0, ErrInvalidLengthMeshsecurity
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMeshsecurity
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMeshsecurity
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMeshsecurity        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMeshsecurity          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMeshsecurity = fmt.Errorf("proto: unexpected end of group")
)
