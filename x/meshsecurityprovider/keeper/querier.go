package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
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

// Params implements the gRPC service handler for querying the mesh-security parameters.
func (q querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params := q.k.GetParams(sdk.UnwrapSDKContext(ctx))
	return &types.QueryParamsResponse{Params: params}, nil
}
