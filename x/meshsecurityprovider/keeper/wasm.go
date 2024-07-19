package keeper

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/contract"
)

func (k Keeper) SendVaultStake(ctx sdk.Context, v contract.StakeMsg) error {
	msg := contract.SudoMsg{
		VaultSudoMsg: &v,
	}

	return k.doSudoCall(ctx, k.GetParams(ctx).GetVaultContractAddress(), msg)
}

func (k Keeper) doSudoCall(ctx sdk.Context, contractAddr sdk.AccAddress, msg contract.SudoMsg) error {
	bz, err := json.Marshal(msg)
	if err != nil {
		return errorsmod.Wrap(err, "marshal sudo msg")
	}
	_, err = k.wasm.Sudo(ctx, contractAddr, bz)
	return err
}
