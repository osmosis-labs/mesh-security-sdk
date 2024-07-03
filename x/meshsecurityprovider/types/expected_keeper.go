package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)


// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	DelegateCoins(ctx sdk.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error
	UndelegateCoins(ctx sdk.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error
}
