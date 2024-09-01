package keeper

import (
	"bytes"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	tmtypes "github.com/cometbft/cometbft/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurityprovider/types"
)

// TODO: testing
// VerifyDoubleVotingEvidence verifies a double voting evidence
// for a given chain id and a validator public key
func (keeper Keeper) VerifyDoubleVotingEvidence(
	evidence tmtypes.DuplicateVoteEvidence,
	chainID string,
	pubkey cryptotypes.PubKey,
) error {
	if pubkey == nil {
		return fmt.Errorf("validator public key cannot be empty")
	}

	// check that the validator address in the evidence is derived from the provided public key
	if !bytes.Equal(pubkey.Address(), evidence.VoteA.ValidatorAddress) {
		return errorsmod.Wrapf(
			types.ErrInvalidDoubleVotingEvidence,
			"public key %s doesn't correspond to the validator address %s in double vote evidence",
			pubkey.String(), evidence.VoteA.ValidatorAddress.String(),
		)
	}

	// Note the age of the evidence isn't checked.

	// height/round/type must be the same
	if evidence.VoteA.Height != evidence.VoteB.Height ||
		evidence.VoteA.Round != evidence.VoteB.Round ||
		evidence.VoteA.Type != evidence.VoteB.Type {
		return errorsmod.Wrapf(
			types.ErrInvalidDoubleVotingEvidence,
			"height/round/type are not the same: %d/%d/%v vs %d/%d/%v",
			evidence.VoteA.Height, evidence.VoteA.Round, evidence.VoteA.Type,
			evidence.VoteB.Height, evidence.VoteB.Round, evidence.VoteB.Type)
	}

	// Addresses must be the same
	if !bytes.Equal(evidence.VoteA.ValidatorAddress, evidence.VoteB.ValidatorAddress) {
		return errorsmod.Wrapf(
			types.ErrInvalidDoubleVotingEvidence,
			"validator addresses do not match: %X vs %X",
			evidence.VoteA.ValidatorAddress,
			evidence.VoteB.ValidatorAddress,
		)
	}

	// BlockIDs must be different
	if evidence.VoteA.BlockID.Equals(evidence.VoteB.BlockID) {
		return errorsmod.Wrapf(
			types.ErrInvalidDoubleVotingEvidence,
			"block IDs are the same (%v) - not a real duplicate vote",
			evidence.VoteA.BlockID,
		)
	}

	va := evidence.VoteA.ToProto()
	vb := evidence.VoteB.ToProto()

	// signatures must be valid
	if !pubkey.VerifySignature(tmtypes.VoteSignBytes(chainID, va), evidence.VoteA.Signature) {
		return fmt.Errorf("verifying VoteA: %w", tmtypes.ErrVoteInvalidSignature)
	}
	if !pubkey.VerifySignature(tmtypes.VoteSignBytes(chainID, vb), evidence.VoteB.Signature) {
		return fmt.Errorf("verifying VoteB: %w", tmtypes.ErrVoteInvalidSignature)
	}

	return nil
}

// TODO: testing
// HandleConsumerDoubleVoting verifies a double voting evidence for a given a consumer chain ID
// and a public key and, if successful, executes the slashing, jailing, and tombstoning of the malicious validator.
func (keeper Keeper) HandleConsumerDoubleVoting(
	ctx sdk.Context,
	evidence *tmtypes.DuplicateVoteEvidence,
	chainID string,
	pubkey cryptotypes.PubKey,
) error {
	// verifies the double voting evidence using the consumer chain public key
	if err := keeper.VerifyDoubleVotingEvidence(*evidence, chainID, pubkey); err != nil {
		return err
	}

	keeper.IteratorProxyStakingContractAddr(ctx, chainID, func(contractAddr string) bool {
		// sudo call `tombstone` to the staking proxy contract
		contractAccAddr, err := sdk.AccAddressFromBech32(contractAddr)
		if err != nil {
			return false
		}
		msg := contract.SudoMsg{
			Tombstoned: &contractAddr,
		}
		err = keeper.doSudoCall(ctx, contractAccAddr, msg)
		if err == nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeSubmitConsumerDoubleVoting,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(types.AttributeKeyContractAddress, contractAddr),
					sdk.NewAttribute(types.AttributeConsumerChainID, chainID),
				),
			)
			keeper.Logger(ctx).Info("Tombstone validator %d in %d by double voting evidence", contractAddr, chainID)
		}
		return false
	})
	return nil
}
