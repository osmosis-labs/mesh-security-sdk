package keeper

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/contract"
)

// caller must ensure gas limits are set proper and handle panics
func (k Keeper) doSudoCall(ctx sdk.Context, contractAddr sdk.AccAddress, msg contract.SudoMsg) error {
	bz, err := json.Marshal(msg)
	if err != nil {
		return errorsmod.Wrap(err, "marshal sudo msg")
	}
	resp, err := k.wasmKeeper.Sudo(ctx, contractAddr, bz)
	fmt.Println("sudo call: ", string(resp))
	fmt.Println("msg: ", string(bz))
	return err
}

// SendJailHandlingMsg send jail handling message to  contract via sudo
func (k Keeper) SendJailHandlingMsg(ctx sdk.Context, contractAddr sdk.AccAddress, jailed []contract.ValidatorAddr, tombstoned []contract.ValidatorAddr) error {
	msg := contract.SudoMsg{
		Jailing: &contract.ValidatorSlash{
			Jailed:     jailed,
			Tombstoned: tombstoned,
		},
	}
	return k.doSudoCall(ctx, contractAddr, msg)
}
