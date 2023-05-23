package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

type Keeper struct {
	storeKey storetypes.StoreKey
	cdc      codec.Codec
	bank     types.XBankKeeper
	staking  types.XStakingKeeper
	// the address capable of executing a MsgUpdateParams message. Typically, this
	// should be the x/gov module account.
	authority string
}

// NewKeeper constructor with vanilla sdk keepers
func NewKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bank types.SDKBankKeeper,
	staking types.SDKStakingKeeper,
	authority string,
) *Keeper {
	return NewKeeperX(cdc, storeKey, NewBankKeeperAdapter(bank), NewStakingKeeperAdapter(staking, bank), authority)
}

// NewKeeperX constructor with extended Osmosis SDK keepers
func NewKeeperX(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bank types.XBankKeeper,
	staking types.XStakingKeeper,
	authority string,
) *Keeper {
	return &Keeper{
		storeKey:  storeKey,
		cdc:       cdc,
		bank:      bank,
		staking:   staking,
		authority: authority,
	}
}

func (k Keeper) HasMaxCapLimit(ctx sdk.Context, actor sdk.AccAddress) bool {
	panic("not implemented, yet")
}
