package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec register types with legacy amino
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDelegate{}, "provider/MsgDelegate", nil)
	cdc.RegisterConcrete(&MsgUndelegate{}, "provider/MsgUndelegate", nil)
	cdc.RegisterConcrete(&MsgSetConsumerCommissionRate{}, "provider/MsgSetConsumerCommissionRate", nil)
}

// RegisterInterfaces register types with interface registry
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgDelegate{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgUndelegate{},
	)
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSetConsumerCommissionRate{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
}
