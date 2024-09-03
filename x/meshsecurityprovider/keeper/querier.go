package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
)

var _ types.QueryServer = &querier{}

type querier struct {
	k *Keeper
}

// NewQuerier constructor
func NewQuerier(k *Keeper) *querier {
	return &querier{k: k}
}

// Params implements the gRPC service handler for querying the mesh-security parameters.
func (q querier) Params(ctx context.Context, req *types.ParamsRequest) (*types.ParamsResponse, error) {
	params := q.k.GetParams(sdk.UnwrapSDKContext(ctx))
	return &types.ParamsResponse{Params: params}, nil
}
