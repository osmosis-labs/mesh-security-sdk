package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	// stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
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

func (k Keeper) Undelegate(ctx sdk.Context, msg types.MsgUndelegate) error {
	amount := wasmkeeper.ConvertSdkCoinToWasmCoin(msg.Amount)
	unbondMsg := contract.VaultCustomMsg{
		Unbond: &contract.UnbondMsg{
			Amount: amount,
		},
	}
	err := k.SendCustomUnbond(ctx, unbondMsg)
	if err != nil {
		return err
	}

	inter, found := k.GetIntermediary(ctx, msg.Amount.Denom)
	if !found {
		return fmt.Errorf("can not found Intermediary by denom %s", msg.Amount.Denom)
	}
	if inter.Token.IsLT(msg.Amount) {
		return fmt.Errorf("failed to undelegate; total inter token %s is smaller than %s", inter.Token, msg.Amount)
	}

	newAmount := inter.Token.Sub(msg.Amount)
	inter.Token = &newAmount

	k.SetIntermediary(ctx, inter)
	return nil
}
