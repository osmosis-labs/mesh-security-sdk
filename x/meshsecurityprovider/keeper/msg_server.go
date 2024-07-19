package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
)

type msgServer struct {
	*Keeper
}

func NewMsgServer(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (k msgServer) SetConsumerCommissionRate(goCtx context.Context, msg *types.MsgSetConsumerCommissionRate) (*types.MsgSetConsumerCommissionRateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: Handle set comissionRate

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSetConsumerCommissionRate,
			sdk.NewAttribute(types.AttributeConsumerChainID, msg.ChainId),
			sdk.NewAttribute(types.AttributeProviderValidatorAddress, msg.ProviderAddr),
			sdk.NewAttribute(types.AttributeConsumerCommissionRate, msg.Rate.String()),
		),
	})

	return &types.MsgSetConsumerCommissionRateResponse{}, nil
}

func (k msgServer) Delegate(goCtx context.Context, msg *types.MsgDelegate) (*types.MsgDelegateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	err := msg.Amount.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to delegate; Validate fail")
	}
	denomDelegate := msg.Amount.Denom

	vaultAdress := sdk.AccAddress(k.Keeper.GetParams(ctx).GetVaultContractAddress())
	delegatorAddress, err := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	if err != nil {
		return nil, err
	}

	balance := k.bank.GetBalance(ctx, delegatorAddress, denomDelegate)

	if balance.IsLT(msg.Amount) {
		return nil, fmt.Errorf("failed to delegate; %s is smaller than %s", balance, msg.Amount)
	}

	if err := k.bank.SendCoins(ctx, delegatorAddress, vaultAdress, sdk.NewCoins([]sdk.Coin{msg.Amount}...)); err != nil {
		return nil, err
	}

	bondDenom := k.Staking.BondDenom(ctx)
	if denomDelegate == bondDenom {
		// local stake
		err = k.Keeper.LocalStake(ctx, msg.Amount, msg.ValidatorAddress)
		if err != nil {
			// if fail refunds
			k.bank.SendCoins(ctx, vaultAdress, delegatorAddress, sdk.NewCoins([]sdk.Coin{msg.Amount}...))
			return nil, err
		}
	} else {
		// remote stake
		err := k.Keeper.RemoteStake(ctx, denomDelegate, msg.Amount)
		if err != nil {
			// if fail refunds
			k.bank.SendCoins(ctx, vaultAdress, delegatorAddress, sdk.NewCoins([]sdk.Coin{msg.Amount}...))
			return nil, err
		}
	}

	depositors, found := k.Keeper.GetDepositors(ctx, msg.DelegatorAddress)
	if !found {
		depositors = types.NewDepositors(msg.DelegatorAddress, []sdk.Coin{msg.Amount})
	} else {
		depositors.Tokens = depositors.Tokens.Add(msg.Amount)
	}
	err = k.Keeper.SetDepositors(ctx, depositors)
	if err != nil {
		return nil, err
	}

	return &types.MsgDelegateResponse{}, nil
}

func (k msgServer) Undelegate(goCtx context.Context, msg *types.MsgUndelegate) (*types.MsgUndelegateResponse, error) {
	return &types.MsgUndelegateResponse{}, nil
}
