package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalid                  = errorsmod.Register(ModuleName, 1, "invalid")
	ErrInvalidConsumerClient    = errorsmod.Register(ModuleName, 16, "ccv channel is not built on correct client")
	ErrConsumerChainNotFound    = errorsmod.Register(ModuleName, 18, "consumer chain not found")
	ErrUnknownConsumerChannelId = errorsmod.Register(ModuleName, 4, "no consumer chain with this channel id")
	ErrUnknownConsumerChainId   = errorsmod.Register(ModuleName, 3, "no consumer chain with this chain id")
)
