package keeper

import (
	"fmt"
	"strconv"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
	cptypes "github.com/osmosis-labs/mesh-security-sdk/x/types"
)

func (k Keeper) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet) error {
	chainID, found := k.GetChannelToChain(ctx, packet.SourceChannel)
	if !found {
		ModuleLogger(ctx).Error("packet timeout, unknown channel:", "channelID", packet.SourceChannel)
		// abort transaction
		return errorsmod.Wrap(
			channeltypes.ErrInvalidChannelState,
			packet.SourceChannel,
		)
	}
	ModuleLogger(ctx).Info("packet timeout, removing the consumer:", "chainID", chainID)
	// stop consumer chain and release unbondings
	return k.StopConsumerChain(ctx, chainID, false)
}

func (k Keeper) StopConsumerChain(ctx sdk.Context, chainID string, closeChan bool) (err error) {
	// check that a client for chainID exists
	if _, found := k.GetConsumerClientId(ctx, chainID); !found {
		return errorsmod.Wrap(types.ErrConsumerChainNotFound,
			fmt.Sprintf("cannot stop non-existent consumer chain: %s", chainID))
	}

	// TODO: Stop consumerchain

	ModuleLogger(ctx).Info("consumer chain removed from provider", "chainID", chainID)

	return nil
}

// OnAcknowledgementPacket handles acknowledgments for sent VSC packets
func (k Keeper) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, ack channeltypes.Acknowledgement) error {
	if err := ack.GetError(); err != "" {
		// The VSC packet data could not be successfully decoded.
		// This should never happen.
		ModuleLogger(ctx).Error(
			"recv ErrorAcknowledgement",
			"channelID", packet.SourceChannel,
			"error", err,
		)
		if chainID, ok := k.GetChannelToChain(ctx, packet.SourceChannel); ok {
			// stop consumer chain and release unbonding
			return k.StopConsumerChain(ctx, chainID, false)
		}
		return errorsmod.Wrapf(types.ErrUnknownConsumerChannelId, "recv ErrorAcknowledgement on unknown channel %s", packet.SourceChannel)
	}
	return nil
}

func (k Keeper) OnRecvSlashPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data cptypes.SlashInfo,
) (cptypes.PacketAckResult, error) {
	chainID, found := k.GetChannelToChain(ctx, packet.DestinationChannel)
	if !found {
		ModuleLogger(ctx).Error("SlashPacket received on unknown channel",
			"channelID", packet.DestinationChannel,
		)
		panic(fmt.Errorf("SlashPacket received on unknown channel %s", packet.DestinationChannel))
	}
	// validate packet data upon receiving
	if err := data.Validate(); err != nil {
		return nil, errorsmod.Wrapf(err, "error validating SlashPacket data")
	}

	if err := k.ValidateSlashPacket(ctx, chainID); err != nil {
		ModuleLogger(ctx).Error("invalid slash packet",
			"error", err.Error(),
			"chainID", chainID,
		)
		return nil, err
	}

	k.HandleSlashPacket(ctx, chainID, data)

	ModuleLogger(ctx).Info("slash packet received and handled",
		"chainID", chainID,
		"consumer val addr", data.Validator,
	)
	return cptypes.SlashPacketHandledResult, nil
}

func (k Keeper) OnRecvBondedPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data *cptypes.ScheduleInfo,
) (cptypes.PacketAckResult, error) {
	intermediary, found := k.GetIntermediary(ctx, data.Denom)
	if !found {
		ModuleLogger(ctx).Error("External Staker not found for validor",
			data.Validator,
		)
		panic(fmt.Errorf("external Staker not found for validor %s", data.Validator))
	}

	if intermediary.Status != types.Bonded {
		intermediary.Status = types.Bonded
	}
	k.SetIntermediary(ctx, intermediary)
	return cptypes.SlashPacketHandledResult, nil
}

func (k Keeper) OnRecvUnbondedPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data *cptypes.ScheduleInfo,
) (cptypes.PacketAckResult, error) {
	intermediary, found := k.GetIntermediary(ctx, data.Denom)
	if !found {
		ModuleLogger(ctx).Error("External Staker not found for validor",
			data.Validator,
		)
		panic(fmt.Errorf("external Staker not found for validor %s", data.Validator))
	}

	if intermediary.Status != types.Unbonded {
		intermediary.Status = types.Unbonded
	}
	k.SetIntermediary(ctx, intermediary)
	return cptypes.SlashPacketHandledResult, nil
}

func (k Keeper) OnRecvJailedPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data *cptypes.ScheduleInfo,
) (cptypes.PacketAckResult, error) {
	intermediary, found := k.GetIntermediary(ctx, data.Denom)
	if !found {
		ModuleLogger(ctx).Error("External Staker not found for validor",
			data.Validator,
		)
		panic(fmt.Errorf("external Staker not found for validor %s", data.Validator))
	}

	if intermediary.IsUnboned() {
		ModuleLogger(ctx).Error("validator %s is unbonded", data.Validator)
		return cptypes.SlashPacketHandledResult, nil
	}

	if intermediary.IsTombstoned() {
		ModuleLogger(ctx).Info(
			"slash packet dropped because validator %s is already tombstoned", data.Validator,
		)
		return cptypes.SlashPacketHandledResult, nil
	}
	if intermediary.IsJailed() {
		ModuleLogger(ctx).Info("validator %s jailed", data.Validator)
		return cptypes.SlashPacketHandledResult, nil
	}
	intermediary.Jailed = true
	k.SetIntermediary(ctx, intermediary)

	return cptypes.SlashPacketHandledResult, nil
}

func (k Keeper) OnRecvTombstonedPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data *cptypes.ScheduleInfo,
) (cptypes.PacketAckResult, error) {
	intermediary, found := k.GetIntermediary(ctx, data.Denom)
	if !found {
		ModuleLogger(ctx).Error("External Staker not found for validor",
			data.Validator,
		)
		panic(fmt.Errorf("external Staker not found for validor %s", data.Validator))
	}

	if intermediary.IsUnboned() {
		ModuleLogger(ctx).Error("validator %s is unbonded", data.Validator)
		return cptypes.SlashPacketHandledResult, nil
	}
	intermediary.Tombstoned = true
	k.SetIntermediary(ctx, intermediary)

	return cptypes.SlashPacketHandledResult, nil
}
func (k Keeper) OnRecvUnjailedPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data *cptypes.ScheduleInfo,
) (cptypes.PacketAckResult, error) {
	intermediary, found := k.GetIntermediary(ctx, data.Denom)
	if !found {
		ModuleLogger(ctx).Error("External Staker not found for validor",
			data.Validator,
		)
		panic(fmt.Errorf("external Staker not found for validor %s", data.Validator))
	}

	if intermediary.IsUnboned() {
		ModuleLogger(ctx).Error("validator %s is unbonded", data.Validator)
		return cptypes.SlashPacketHandledResult, nil
	}
	if !intermediary.Jailed {
		ModuleLogger(ctx).Error("validator %s is not jailed", data.Validator)
		return cptypes.SlashPacketHandledResult, nil
	}

	intermediary.Jailed = false
	k.SetIntermediary(ctx, intermediary)

	return cptypes.SlashPacketHandledResult, nil
}
func (k Keeper) OnRecvModifiedPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	data *cptypes.ScheduleInfo,
) (cptypes.PacketAckResult, error) {

	return cptypes.SlashPacketHandledResult, nil
}

func (k Keeper) HandleSlashPacket(ctx sdk.Context, chainID string, data cptypes.SlashInfo) {
	totalSlashAmount, err := sdk.ParseCoinNormalized(data.TotalSlashAmount)
	if err != nil {
		ModuleLogger(ctx).Error("Handle slash packet fail: ParseCoinsNormalized fail")
		return
	}
	denom := totalSlashAmount.Denom
	intermediary, found := k.GetIntermediary(ctx, denom)
	if !found {
		ModuleLogger(ctx).Error("External Staker not found for validor",
			data.Validator,
		)
		panic(fmt.Errorf("external Staker not found for validor %s", data.Validator))
	}

	if intermediary.IsUnboned() {
		ModuleLogger(ctx).Error("validator is unbonded")
		return
	}

	if intermediary.IsTombstoned() {
		ModuleLogger(ctx).Info(
			"slash packet dropped because validator is already tombstoned",
		)
		return
	}
	if intermediary.IsJailed() {
		ModuleLogger(ctx).Info("validator jailed")
		return
	}

	intermediary.Jailed = true
	slashRatio := sdk.MustNewDecFromStr(data.SlashFraction)
	amountSlash := slashRatio.MulInt(intermediary.Token.Amount).TruncateInt()
	newAmount := sdk.NewCoin(denom, amountSlash)
	intermediary.Token = &newAmount
	k.SetIntermediary(ctx, intermediary)

	k.iterateDepositors(ctx, func(depositors types.Depositors) bool {
		amout := depositors.Tokens.AmountOf(denom)
		if amout.GT(sdk.ZeroInt()) {
			amountSlash = slashRatio.MulInt(amout).TruncateInt()
			tokenSlash := sdk.NewCoin(denom, amountSlash)
			newTokens := depositors.Tokens.Sub(tokenSlash)
			depositors.Tokens = newTokens

			k.SetDepositors(ctx, depositors)
		}
		return false
	})
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeExecuteConsumerChainSlash,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
			sdk.NewAttribute(types.AttributeInfractionHeight, strconv.FormatInt(data.InfractionHeight, 10)),
			sdk.NewAttribute(types.AttributeConsumerValidator, data.Validator),
		),
	)
}

func (k Keeper) ValidateSlashPacket(ctx sdk.Context, chainID string,
) error {
	_, found := k.GetInitChainHeight(ctx, chainID)
	// return error if we cannot find infraction height matching the validator update id
	if !found {
		return fmt.Errorf("cannot find infraction height matching "+
			"for chain %s", chainID)
	}

	return nil
}
