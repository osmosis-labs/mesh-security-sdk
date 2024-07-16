package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	// stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/provider/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/provider/types"
)

func (k Keeper) LocalStake(ctx sdk.Context, token sdk.Coin, valAddr string) error {
	amount := wasmkeeper.ConvertSdkCoinToWasmCoin(token)
	providerStakeMsg := contract.StakeMsg{
		StakeLocal: contract.StakeLocal{
			Amount: amount,
		},
	}

	err := k.SendVaultStake(ctx, providerStakeMsg)
	if err != nil {
		return err
	}

	contractAddr := k.GetContractWithNativeDenom(ctx, token.Denom)
	inter, found := k.GetIntermediary(ctx, token.Denom)
	if !found {
		inter = types.NewIntermediary(valAddr, ctx.ChainID(), contractAddr.String(), false, false, types.Bonded, &token)
	} else {
		newAmout := inter.Token.Add(token)
		inter.Token = &newAmout
	}
	k.SetIntermediary(ctx, inter)
	return nil
}

func (k Keeper) RemoteStake(ctx sdk.Context, denomDelegate string, token sdk.Coin) error {
	contractAddr := k.GetContractWithNativeDenom(ctx, denomDelegate)

	amount := wasmkeeper.ConvertSdkCoinToWasmCoin(token)

	providerStakeMsg := contract.StakeMsg{
		StakeRemote: contract.StakeRemote{
			Amount:   amount,
			Contract: contractAddr.String(),
		},
	}
	err := k.SendVaultStake(ctx, providerStakeMsg)
	if err != nil {
		return err
	}

	// TODO: validator addresss on consumer chain
	inter, found := k.GetIntermediary(ctx, token.Denom)
	if !found {
		inter = types.NewIntermediary("", ctx.ChainID(), contractAddr.String(), false, false, types.Bonded, &token)
	} else {
		newAmout := inter.Token.Add(token)
		inter.Token = &newAmout
	}
	k.SetIntermediary(ctx, inter)
	return nil
}
