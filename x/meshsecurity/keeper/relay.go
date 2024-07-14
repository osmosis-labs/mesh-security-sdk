package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	ctypes "github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
	"github.com/osmosis-labs/mesh-security-sdk/x/types"
)

func (k Keeper) SendPackets(ctx sdk.Context) {
	channelID, ok := k.GetProviderChannel(ctx)
	if !ok {
		return
	}

	// pending := k.GetAllPendingPacketsWithIdx(ctx)
	ConsumerPackets := []types.ConsumerPacketData{}
	k.iteratePipedValsetOperations(ctx, func(valAddress sdk.ValAddress, op types.PipedValsetOperation, slashInfo *types.SlashInfo) bool {
		newPacket := types.ConsumerPacketData{
			Type: op,
			Data: &types.ConsumerPacketData_SlashPacketData{
				SlashPacketData: slashInfo,
			},
		}
		ConsumerPackets = append(ConsumerPackets, newPacket)

		return false
	})

	for _, s := range ConsumerPackets {
		// Send packet over IBC
		err := ctypes.SendIBCPacket(
			ctx,
			k.scopedKeeper,
			k.channelKeeper,
			channelID,             // source channel id
			ctypes.ConsumerPortID, // source port id
			s.GetBytes(),
			k.GetParams(ctx).GetTimeoutPeriod(),
		)
		if err != nil {
			if clienttypes.ErrClientNotActive.Is(err) {
				ModuleLogger(ctx).Info("IBC client is expired, cannot send IBC packet; leaving packet data stored:")
				break
			}
			ModuleLogger(ctx).Error("cannot send IBC packet; leaving packet data stored:", "err", err.Error())
			break
		}
	}
}

func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, ack channeltypes.Acknowledgement) error {
	if res := ack.GetResult(); res != nil {
		if len(res) != 1 {
			return fmt.Errorf("acknowledgement result length must be 1, got %d", len(res))
		}

		consumerPacket, err := types.UnmarshalConsumerPacketData(packet.GetData())
		if err != nil {
			return err
		}
		if consumerPacket.Type != types.PipedValsetOperation_VALIDATOR_SLASHED {
			return nil
		}
		k.ClearPipedValsetOperations(ctx)
	}

	if err := ack.GetError(); err != "" {
		// Reasons for ErrorAcknowledgment
		//  - packet data could not be successfully decoded
		//  - invalid Slash packet
		// None of these should ever happen.
		ModuleLogger(ctx).Error(
			"recv ErrorAcknowledgement",
			"channel", packet.SourceChannel,
			"error", err,
		)
		// Initiate ChanCloseInit using packet source (non-counterparty) port and channel
		err := k.ChanCloseInit(ctx, packet.SourcePort, packet.SourceChannel)
		if err != nil {
			return fmt.Errorf("ChanCloseInit(%s) failed: %s", packet.SourceChannel, err.Error())
		}
		// check if there is an established CCV channel to provider
		channelID, found := k.GetProviderChannel(ctx)
		if !found {
			return errorsmod.Wrapf(ctypes.ErrNoProposerChannelId, "recv ErrorAcknowledgement on non-established channel %s", packet.SourceChannel)
		}
		if channelID != packet.SourceChannel {
			// Close the established CCV channel as well
			return k.ChanCloseInit(ctx, types.ConsumerPortID, channelID)
		}
	}
	return nil
}
