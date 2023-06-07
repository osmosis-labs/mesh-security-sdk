package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
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

// VirtualStakingMaxCapLimit returns limit amount for given contract. Returns 0 amount for unknown addresses
func (g querier) VirtualStakingMaxCapLimit(goCtx context.Context, req *types.QueryVirtualStakingMaxCapLimitRequest) (*types.QueryVirtualStakingMaxCapLimitResponse, error) {
	acc, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, errorsmod.Wrap(err, "contract")
	}
	return &types.QueryVirtualStakingMaxCapLimitResponse{Cap: g.k.GetMaxCapLimit(sdk.UnwrapSDKContext(goCtx), acc), Delegated: g.k.GetTotalDelegated(sdk.UnwrapSDKContext(goCtx), acc)}, nil
}

// VirtualStakingMaxCapLimits returns limit amount for all the contracts.
func (g querier) VirtualStakingMaxCapLimits(goCtx context.Context, req *types.QueryVirtualStakingMaxCapLimitsRequest) (*types.QueryVirtualStakingMaxCapLimitsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	rsp := types.QueryVirtualStakingMaxCapLimitsResponse{}
	g.k.IterateMaxCapLimit(ctx, func(addr sdk.AccAddress, maxCap math.Int) bool {
		info := types.VirtualStakingMaxCapInfo{
			Contract:  addr.String(),
			Delegated: g.k.GetTotalDelegated(ctx, addr),
			Cap:       sdk.NewCoin(g.k.staking.BondDenom(ctx), maxCap),
		}

		rsp.MaxCapInfos = append(rsp.MaxCapInfos, info)
		return false
	})

	return &rsp, nil
}

// Params implements the gRPC service handler for querying the mesh-security parameters.
func (q querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params := q.k.GetParams(sdk.UnwrapSDKContext(ctx))
	return &types.QueryParamsResponse{Params: params}, nil
}
