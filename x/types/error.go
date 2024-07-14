package types

import (
	errorsmod "cosmossdk.io/errors"
)

// meshsecurity sentinel errors
var (
	ErrInvalidVersion     = errorsmod.Register(ModuleName, 2, "invalid meshsecurity version")
	ErrInvalidChannelFlow = errorsmod.Register(ModuleName, 3, "invalid message sent to channel end")
	ErrDuplicateChannel   = errorsmod.Register(ModuleName, 5, "meshsecurity channel already exists")
	ErrClientNotFound     = errorsmod.Register(ModuleName, 10, "client not found")
)
