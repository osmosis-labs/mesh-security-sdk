package keeper

import (
	"context"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	if err := m.k.SetMaxCapLimit(sdk.UnwrapSDKContext(goCtx), acc, req.MaxCap); err != nil {
		return nil, err
	}
	return &types.MsgSetVirtualStakingMaxCapResponse{}, nil
}
