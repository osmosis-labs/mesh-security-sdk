package keeper

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
)

// Rebalance send rebalance message to virtual staking contract
func (k Keeper) Rebalance(ctx sdk.Context, addr sdk.AccAddress) error {
	msg := contract.SudoMsg{
		Rebalance: &struct{}{},
	}
	bz, err := json.Marshal(msg)
	if err != nil {
		return errorsmod.Wrap(err, "marshal sudo msg")
	}
	_, err = k.sudoer.Sudo(ctx, addr, bz)
	return err
}
