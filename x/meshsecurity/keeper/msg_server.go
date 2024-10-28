package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	k *Keeper
}

// NewMsgServer constructor
func NewMsgServer(k *Keeper) *msgServer {
	return &msgServer{k: k}
}

// SetVirtualStakingMaxCap sets a new max cap limit for virtual staking
func (m msgServer) SetVirtualStakingMaxCap(goCtx context.Context, req *types.MsgSetVirtualStakingMaxCap) (*types.MsgSetVirtualStakingMaxCapResponse, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}

	if authority := m.k.GetAuthority(); authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", authority, req.Authority)
	}

	acc, err := sdk.AccAddressFromBech32(req.Contract)
	if err != nil {
		return nil, errorsmod.Wrap(err, "contract")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.k.SetMaxCapLimit(ctx, acc, req.MaxCap); err != nil {
		return nil, err
	}
	if !m.k.HasScheduledTask(ctx, types.SchedulerTaskHandleEpoch, acc, true) {
		if err := m.k.ScheduleRegularHandleEpochTask(ctx, acc); err != nil {
			return nil, errorsmod.Wrap(err, "schedule regular rebalance task")
		}
		return &types.MsgSetVirtualStakingMaxCapResponse{}, nil
	}
	return &types.MsgSetVirtualStakingMaxCapResponse{}, nil
}

// SetPriceFeedContract sets the price feed contract to the chain to trigger handle epoch task
func (m msgServer) SetPriceFeedContract(goCtx context.Context, req *types.MsgSetPriceFeedContract) (*types.MsgSetPriceFeedContractResponse, error) {
	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}

	if authority := m.k.GetAuthority(); authority != req.Authority {
		return nil, govtypes.ErrInvalidSigner.Wrapf("invalid authority; expected %s, got %s", authority, req.Authority)
	}

	acc, err := sdk.AccAddressFromBech32(req.Contract)
	if err != nil {
		return nil, errorsmod.Wrap(err, "contract")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	if !m.k.HasScheduledTask(ctx, types.SchedulerTaskHandleEpoch, acc, true) {
		if err := m.k.ScheduleRegularHandleEpochTask(ctx, acc); err != nil {
			return nil, errorsmod.Wrap(err, "schedule regular rebalance task")
		}
		return &types.MsgSetPriceFeedContractResponse{}, nil
	} else {
		return nil, types.ErrDuplicate
	}
}
