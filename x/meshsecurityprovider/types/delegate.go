package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func (e Intermediary) IsTombstoned() bool {
	return e.Tombstoned
}

func (e Intermediary) IsUnboned() bool {
	return e.Status == Unbonded
}

func (e Intermediary) IsJailed() bool {
	return e.Jailed
}

func NewIntermediary(consumerValidator, chainID, contractAddress string, jailed, tombstoned bool, status BondStatus, token *sdk.Coin) Intermediary {
	return Intermediary{
		ConsumerValidator: consumerValidator,
		ContractAddress:   contractAddress,
		ChainId:           chainID,
		Jailed:            jailed,
		Tombstoned:        tombstoned,
		Status:            status,
		Token:             token,
	}
}

func NewDepositors(address string, tokens sdk.Coins) Depositors {
	return Depositors{
		Address: address,
		Tokens:  tokens,
	}
}
