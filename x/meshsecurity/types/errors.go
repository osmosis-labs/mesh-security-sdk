package types

import (
	errorsmod "cosmossdk.io/errors"
)

var ErrInvalid = errorsmod.Register(ModuleName, 1, "invalid")
