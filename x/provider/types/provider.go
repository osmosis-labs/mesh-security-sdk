package types

import (
	crypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
)

type ValidatorConsumerPubKey struct {
	ChainId      string
	ProviderAddr []byte
	ConsumerKey  *crypto.PublicKey
}

type ValidatorByConsumerAddr struct {
	ChainId      string
	ConsumerAddr []byte
	ProviderAddr []byte
}

type ConsumerAddrsToPrune struct {
	ChainId       string
	VscId         uint64
	ConsumerAddrs *AddressList
}

// AddressList contains a list of consensus addresses
type AddressList struct {
	Addresses [][]byte
}
