package keeper

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
)

// SendHandleEpoch send epoch handling message to virtual staking contract via sudo
func (k Keeper) SendHandleEpoch(ctx sdk.Context, contractAddr sdk.AccAddress) error {
	msg := contract.SudoMsg{
		HandleEpoch: &struct{}{},
	}
	return k.doSudoCall(ctx, contractAddr, msg)
}

// SendValsetUpdate submit the valset update report to the virtual staking contract via sudo
func (k Keeper) SendValsetUpdate(ctx sdk.Context, contractAddr sdk.AccAddress, v contract.ValsetUpdate) error {
	msg := contract.SudoMsg{
		ValsetUpdate: &v,
	}
	return k.doSudoCall(ctx, contractAddr, msg)
}

// caller must ensure gas limits are set proper and handle panics
func (k Keeper) doSudoCall(ctx sdk.Context, contractAddr sdk.AccAddress, msg contract.SudoMsg) error {
	bz, err := json.Marshal(msg)
	if err != nil {
		return errorsmod.Wrap(err, "marshal sudo msg")
	}
	_, err = k.wasm.Sudo(ctx, contractAddr, bz)
	return err
}
