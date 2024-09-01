package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidDoubleVotingEvidence = errorsmod.Register(ModuleName, 1, "invalid consumer double voting evidence")
)
