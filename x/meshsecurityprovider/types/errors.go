package types

import (
	errorsmod "cosmossdk.io/errors"
)

var (
	ErrInvalidDoubleVotingEvidence        = errorsmod.Register(ModuleName, 1, "invalid consumer double voting evidence")
	ErrUnsupportedCounterpartyClientTypes = errorsmod.Register(ModuleName, 2, "invalid counterparty client type, only support tendermint light client")
)
