package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

var _ types.QueryServer = &querier{}

type querier struct {
	cdc codec.Codec
	k   *Keeper
}

// NewQuerier constructor
func NewQuerier(cdc codec.Codec, k *Keeper) *querier {
	return &querier{cdc: cdc, k: k}
}

// VirtualStakingMaxCap returns limit amount for given contract. Returns 0 amount for unknown addresses
func (g querier) VirtualStakingMaxCap(goCtx context.Context, req *types.QueryVirtualStakingMaxCapRequest) (*types.QueryVirtualStakingMaxCapResponse, error) {
	acc, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errorsmod.Wrap(err, "contract")
	}
	return &types.QueryVirtualStakingMaxCapResponse{Limit: g.k.GetMaxCapLimit(sdk.UnwrapSDKContext(goCtx), acc)}, nil
}
