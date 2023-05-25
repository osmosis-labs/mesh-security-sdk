package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalid        = errorsmod.Register(ModuleName, 1, "invalid")
	ErrMaxCapExceeded = errorsmod.Register(ModuleName, 2, "max cap exceeded")
)
