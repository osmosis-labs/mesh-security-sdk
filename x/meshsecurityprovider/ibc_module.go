package meshsecurityprovider

import (
	"fmt"
	// "strconv"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	// transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/keeper"
	types "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
	cptypes "github.com/osmosis-labs/mesh-security-sdk/x/types"
)

// OnChanOpenInit implements the IBCModule interface
func (am AppModule) OnChanOpenInit(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string,
	channelID string,
	channelCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return version, errorsmod.Wrap(cptypes.ErrInvalidChannelFlow, "channel handshake must be initiated by consumer chain")
}

// OnChanOpenTry implements the IBCModule interface

func (am AppModule) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID,
	channelID string,
	chanCap *capabilitytypes.Capability,
	counterparty channeltypes.Counterparty,
	counterpartyVersion string,
) (metadata string, err error) {
	if counterparty.PortId != cptypes.ConsumerPortID {
		return "", errorsmod.Wrapf(porttypes.ErrInvalidPort,
			"invalid counterparty port: %s, expected %s", counterparty.PortId, cptypes.ConsumerPortID)
	}

	// ensure the counter party version matches the expected version
	if counterpartyVersion != cptypes.Version {
		return "", errorsmod.Wrapf(
			cptypes.ErrInvalidVersion, "invalid counterparty version: got: %s, expected %s",
			counterpartyVersion, cptypes.Version)
	}

	// Claim channel capability
	if err := am.k.ClaimCapability(
		ctx, chanCap, host.ChannelCapabilityPath(portID, channelID),
	); err != nil {
		return "", err
	}

	if err := am.k.VerifyConsumerChain(
		ctx, channelID, connectionHops,
	); err != nil {
		return "", err
	}
	// TODO: ConsumerRewards
	return "", nil
}

// OnChanOpenAck implements the IBCModule interface
func (am AppModule) OnChanOpenAck(
	ctx sdk.Context,
	portID,
	channelID string,
	_ string, // Counter party channel ID is unused per spec
	counterpartyMetadata string,
) error {
	return errorsmod.Wrap(cptypes.ErrInvalidChannelFlow, "channel handshake must be initiated by consumer chain")
}

// OnChanOpenConfirm implements the IBCModule interface
func (am AppModule) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	err := am.k.SetConsumerChain(ctx, channelID)
	if err != nil {
		return err
	}
	return nil
}

// OnChanCloseInit implements the IBCModule interface
func (am AppModule) OnChanCloseInit(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "user cannot close channel")
}

// OnChanCloseConfirm implements the IBCModule interface
func (am AppModule) OnChanCloseConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	return nil
}

// OnRecvPacket implements the IBCModule interface. A successful acknowledgement
// is returned if the packet data is successfully decoded and the receive application
// logic returns without error.
func (am AppModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	_ sdk.AccAddress,
) ibcexported.Acknowledgement {
	logger := keeper.ModuleLogger(ctx)
	ack := channeltypes.NewResultAcknowledgement([]byte{byte(1)})

	var ackErr error
	consumerPacket, err := UnmarshalConsumerPacket(packet)
	if err != nil {
		ackErr = errorsmod.Wrapf(sdkerrors.ErrInvalidType, "cannot unmarshal ConsumerPacket data")
		logger.Error(fmt.Sprintf("%s sequence %d", ackErr.Error(), packet.Sequence))
		ack = channeltypes.NewErrorAcknowledgement(ackErr)
	}

	eventAttributes := []sdk.Attribute{
		sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
	}

	if ack.Success() {
		var err error

		switch consumerPacket.Type {
		case cptypes.PipedValsetOperation_VALIDATOR_BONDED:
			var ackResult cptypes.PacketAckResult
			data := consumerPacket.GetSchedulePacketData()
			ackResult, err = am.k.OnRecvBondedPacket(ctx, packet, data)
			if err == nil {
				ack = channeltypes.NewResultAcknowledgement(ackResult)
			}

		case cptypes.PipedValsetOperation_VALIDATOR_UNBONDED:
			var ackResult cptypes.PacketAckResult
			data := consumerPacket.GetSchedulePacketData()
			ackResult, err = am.k.OnRecvUnbondedPacket(ctx, packet, data)
			if err == nil {
				ack = channeltypes.NewResultAcknowledgement(ackResult)
			}
		case cptypes.PipedValsetOperation_VALIDATOR_JAILED:
			var ackResult cptypes.PacketAckResult
			data := consumerPacket.GetSchedulePacketData()
			ackResult, err = am.k.OnRecvJailedPacket(ctx, packet, data)
			if err == nil {
				ack = channeltypes.NewResultAcknowledgement(ackResult)
			}
		case cptypes.PipedValsetOperation_VALIDATOR_TOMBSTONED:
			var ackResult cptypes.PacketAckResult
			data := consumerPacket.GetSchedulePacketData()
			ackResult, err = am.k.OnRecvTombstonedPacket(ctx, packet, data)
			if err == nil {
				ack = channeltypes.NewResultAcknowledgement(ackResult)
			}
		case cptypes.PipedValsetOperation_VALIDATOR_UNJAILED:
			var ackResult cptypes.PacketAckResult
			data := consumerPacket.GetSchedulePacketData()
			ackResult, err = am.k.OnRecvUnjailedPacket(ctx, packet, data)
			if err == nil {
				ack = channeltypes.NewResultAcknowledgement(ackResult)
			}
		case cptypes.PipedValsetOperation_VALIDATOR_MODIFIED:
			var ackResult cptypes.PacketAckResult
			data := consumerPacket.GetSchedulePacketData()
			ackResult, err = am.k.OnRecvModifiedPacket(ctx, packet, data)
			if err == nil {
				ack = channeltypes.NewResultAcknowledgement(ackResult)
			}
		case cptypes.PipedValsetOperation_VALIDATOR_SLASHED:
			var ackResult cptypes.PacketAckResult
			data := consumerPacket.GetSlashPacketData()
			ackResult, err = am.k.OnRecvSlashPacket(ctx, packet, *data)
			if err == nil {
				ack = channeltypes.NewResultAcknowledgement(ackResult)
			}
		default:
			err = fmt.Errorf("invalid consumer packet type: %q", consumerPacket.Type)
		}

		if err != nil {
			ack = channeltypes.NewErrorAcknowledgement(err)
			ackErr = err
			logger.Error(fmt.Sprintf("%s sequence %d", ackErr.Error(), packet.Sequence))
		}
	}

	eventAttributes = append(eventAttributes, sdk.NewAttribute(cptypes.AttributeKeyAckSuccess, fmt.Sprintf("%t", ack.Success())))
	if ackErr != nil {
		eventAttributes = append(eventAttributes, sdk.NewAttribute(cptypes.AttributeKeyAckError, ackErr.Error()))
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			cptypes.EventTypePacket,
			eventAttributes...,
		),
	)

	// NOTE: acknowledgement will be written synchronously during IBC handler execution.
	return ack
}

// OnAcknowledgementPacket implements the IBCModule interface
func (am AppModule) OnAcknowledgementPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	acknowledgement []byte,
	_ sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal provider packet acknowledgement: %v", err)
	}

	if err := am.k.OnAcknowledgementPacket(ctx, packet, ack); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			cptypes.EventTypePacket,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(cptypes.AttributeKeyAck, ack.String()),
		),
	)

	switch resp := ack.Response.(type) {
	case *channeltypes.Acknowledgement_Result:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				cptypes.EventTypePacket,
				sdk.NewAttribute(cptypes.AttributeKeyAckSuccess, string(resp.Result)),
			),
		)
	case *channeltypes.Acknowledgement_Error:
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				cptypes.EventTypePacket,
				sdk.NewAttribute(cptypes.AttributeKeyAckError, resp.Error),
			),
		)
	}

	return nil
}

// OnTimeoutPacket implements the IBCModule interface
func (am AppModule) OnTimeoutPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	_ sdk.AccAddress,
) error {
	if err := am.k.OnTimeoutPacket(ctx, packet); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			cptypes.EventTypeTimeout,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
		),
	)

	return nil
}

func UnmarshalConsumerPacket(packet channeltypes.Packet) (consumerPacket cptypes.ConsumerPacketData, err error) {
	return cptypes.UnmarshalConsumerPacketData(packet.GetData())
}
