package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalid                              = errorsmod.Register(ModuleName, 1, "invalid")
	ErrMaxCapExceeded                       = errorsmod.Register(ModuleName, 2, "max cap exceeded")
	ErrUnsupported                          = errorsmod.Register(ModuleName, 3, "unsupported")
	ErrUnknown                              = errorsmod.Register(ModuleName, 4, "unknown")
	ErrNoProposerChannelId                  = errorsmod.Register(ModuleName, 5, "no established meshsecurity channel")
	ErrConsumerRewardDenomAlreadyRegistered = errorsmod.Register(ModuleName, 6, "consumer reward denom already registered")
)
