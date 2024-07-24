package types

import (
	"strings"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewMsgDelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) *MsgDelegate {
	return &MsgDelegate{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
		Amount:           amount,
	}
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgDelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for MsgSoftwareUpgrade.
func (msg MsgDelegate) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic validate basic constraints
func (msg MsgDelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return errorsmod.Wrap(err, "contract")
	}
	if err := msg.Amount.Validate(); err != nil {
		return errorsmod.Wrap(err, "max cap")
	}
	return nil
}

func NewMsgUndelegate(delAddr sdk.AccAddress, valAddr sdk.ValAddress, amount sdk.Coin) *MsgUndelegate {
	return &MsgUndelegate{
		DelegatorAddress: delAddr.String(),
		ValidatorAddress: valAddr.String(),
		Amount:           amount,
	}
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUndelegate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for MsgSoftwareUpgrade.
func (msg MsgUndelegate) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.DelegatorAddress)
	return []sdk.AccAddress{addr}
}

// ValidateBasic validate basic constraints
func (msg MsgUndelegate) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.DelegatorAddress); err != nil {
		return sdkerrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}
	if _, err := sdk.ValAddressFromBech32(msg.ValidatorAddress); err != nil {
		return errorsmod.Wrap(err, "contract")
	}
	if err := msg.Amount.Validate(); err != nil {
		return errorsmod.Wrap(err, "max cap")
	}
	return nil
}

// NewMsgSetConsumerCommissionRate creates a new MsgSetConsumerCommissionRate msg instance.
func NewMsgSetConsumerCommissionRate(chainID string, commission sdk.Dec, providerValidatorAddress sdk.ValAddress) *MsgSetConsumerCommissionRate {
	return &MsgSetConsumerCommissionRate{
		ChainId:      chainID,
		Rate:         commission,
		ProviderAddr: providerValidatorAddress.String(),
	}
}

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgSetConsumerCommissionRate) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for MsgSoftwareUpgrade.
func (msg MsgSetConsumerCommissionRate) GetSigners() []sdk.AccAddress {
	valAddr, err := sdk.ValAddressFromBech32(msg.ProviderAddr)
	if err != nil {
		// same behavior as in cosmos-sdk
		panic(err)
	}
	return []sdk.AccAddress{valAddr.Bytes()}
}

// ValidateBasic validate basic constraints
func (msg MsgSetConsumerCommissionRate) ValidateBasic() error {
	if strings.TrimSpace(msg.ChainId) == "" {
		panic("chainId cannot be blank")
	}

	if 128 < len(msg.ChainId) {
		panic("chainId cannot exceed 128 length")
	}
	_, err := sdk.ValAddressFromBech32(msg.ProviderAddr)
	if err != nil {
		panic(err)
	}

	if msg.Rate.IsNegative() || msg.Rate.GT(sdk.OneDec()) {
		panic("consumer commission rate should be in the range [0, 1]")
	}

	return nil
}
