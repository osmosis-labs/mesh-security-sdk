package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalid        = errorsmod.Register(ModuleName, 1, "invalid")
	ErrMaxCapExceeded = errorsmod.Register(ModuleName, 2, "max cap exceeded")
	ErrUnsupported    = errorsmod.Register(ModuleName, 3, "unsupported")
	ErrUnknown        = errorsmod.Register(ModuleName, 4, "unknown")
)
