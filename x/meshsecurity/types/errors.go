package types

import (
	errorsmod "cosmossdk.io/errors"
)

var ErrMaxCapExceeded = errorsmod.Register(ModuleName, 1, "max cap exceeded")
