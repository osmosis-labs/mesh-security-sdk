package types

import (
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

var _ sdk.Msg = &MsgUpdateParams{}

const (
	TypeMsgUpdateParams               = "update-params"
	TypeMsgSubmitConsumerMisbehaviour = "submit-consumer-misbehaviour"
	TypeMsgSubmitConsumerDoubleVoting = "submit-consumer-double-voting"
)

func (msg MsgUpdateParams) Route() string { return ModuleName }
func (msg MsgUpdateParams) Type() string  { return TypeMsgUpdateParams }
func (msg MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errorsmod.Wrap(err, "invalid authority address")
	}
	return msg.Params.Validate()
}

func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

func (msg MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

func NewMsgSubmitConsumerMisbehaviour(submitter sdk.AccAddress, misbehaviour *ibctmtypes.Misbehaviour) (*MsgSubmitConsumerMisbehaviour, error) {
	return &MsgSubmitConsumerMisbehaviour{Submitter: submitter.String(), Misbehaviour: misbehaviour}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgSubmitConsumerMisbehaviour) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgSubmitConsumerMisbehaviour) Type() string {
	return TypeMsgSubmitConsumerMisbehaviour
}

// Type implements the sdk.Msg interface.
func (msg MsgSubmitConsumerMisbehaviour) ValidateBasic() error {
	if msg.Submitter == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msg.Submitter)
	}

	if err := msg.Misbehaviour.ValidateBasic(); err != nil {
		return err
	}
	return nil
}

// Type implements the sdk.Msg interface.
func (msg MsgSubmitConsumerMisbehaviour) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// Type implements the sdk.Msg interface.
func (msg MsgSubmitConsumerMisbehaviour) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		// same behavior as in cosmos-sdk
		panic(err)
	}
	return []sdk.AccAddress{addr}
}

func NewMsgSubmitConsumerDoubleVoting(submitter sdk.AccAddress, ev *tmproto.DuplicateVoteEvidence, header *ibctmtypes.Header) (*MsgSubmitConsumerDoubleVoting, error) {
	return &MsgSubmitConsumerDoubleVoting{Submitter: submitter.String(), DuplicateVoteEvidence: ev, InfractionBlockHeader: header}, nil
}

// Route implements the sdk.Msg interface.
func (msg MsgSubmitConsumerDoubleVoting) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (msg MsgSubmitConsumerDoubleVoting) Type() string {
	return TypeMsgSubmitConsumerDoubleVoting
}

// Type implements the sdk.Msg interface.
func (msg MsgSubmitConsumerDoubleVoting) ValidateBasic() error {
	if msg.Submitter == "" {
		return errorsmod.Wrap(sdkerrors.ErrInvalidAddress, msg.Submitter)
	}
	if msg.DuplicateVoteEvidence == nil {
		return fmt.Errorf("double voting evidence cannot be nil")
	}

	if msg.InfractionBlockHeader == nil {
		return fmt.Errorf("infraction block header cannot be nil")
	}

	if msg.InfractionBlockHeader.SignedHeader == nil {
		return fmt.Errorf("signed header in infraction block header cannot be nil")
	}

	if msg.InfractionBlockHeader.SignedHeader.Header == nil {
		return fmt.Errorf("invalid signed header in infraction block header, 'SignedHeader.Header' is nil")
	}

	if msg.InfractionBlockHeader.ValidatorSet == nil {
		return fmt.Errorf("invalid infraction block header, validator set is nil")
	}

	return nil
}

// Type implements the sdk.Msg interface.
func (msg MsgSubmitConsumerDoubleVoting) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// Type implements the sdk.Msg interface.
func (msg MsgSubmitConsumerDoubleVoting) GetSigners() []sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(msg.Submitter)
	if err != nil {
		// same behavior as in cosmos-sdk
		panic(err)
	}
	return []sdk.AccAddress{addr}
}
