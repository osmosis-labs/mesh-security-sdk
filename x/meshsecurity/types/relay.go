package types

import (
	errorsmod "cosmossdk.io/errors"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	host "github.com/cosmos/ibc-go/v7/modules/core/24-host"
)

// SendIBCPacket sends an IBC packet with packetData
// over the source channelID and portID
func SendIBCPacket(
	ctx sdk.Context,
	scopedKeeper ScopedKeeper,
	channelKeeper ChannelKeeper,
	sourceChannelID string,
	sourcePortID string,
	packetData []byte,
	timeoutPeriod time.Duration,
) error {
	_, ok := channelKeeper.GetChannel(ctx, sourcePortID, sourceChannelID)
	if !ok {
		return errorsmod.Wrapf(channeltypes.ErrChannelNotFound, "channel not found for channel ID: %s", sourceChannelID)
	}
	channelCap, ok := scopedKeeper.GetCapability(ctx, host.ChannelCapabilityPath(sourcePortID, sourceChannelID))
	if !ok {
		return errorsmod.Wrap(channeltypes.ErrChannelCapabilityNotFound, "module does not own channel capability")
	}

	_, err := channelKeeper.SendPacket(ctx,
		channelCap,
		sourcePortID,
		sourceChannelID,
		clienttypes.Height{}, //  timeout height disabled
		uint64(ctx.BlockTime().Add(timeoutPeriod).UnixNano()), // timeout timestamp
		packetData,
	)
	return err
}
