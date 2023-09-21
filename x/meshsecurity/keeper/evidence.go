package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

var _ evidencetypes.SlashingKeeper = EvidenceDecorator{}

// EvidenceDecorator to capture TompStone events
type EvidenceDecorator struct {
	evidencetypes.SlashingKeeper
	stakingKeeper types.SDKStakingKeeper
	k             *Keeper
}

// CaptureTombstoneDecorator constructor
func CaptureTombstoneDecorator(k *Keeper, slashingKeeper evidencetypes.SlashingKeeper, stakingKeeper types.SDKStakingKeeper) *EvidenceDecorator {
	return &EvidenceDecorator{SlashingKeeper: slashingKeeper, stakingKeeper: stakingKeeper, k: k}
}

// Tombstone is executed in endblocker by evidence module
func (e EvidenceDecorator) Tombstone(ctx sdk.Context, address sdk.ConsAddress) {
	v, ok := e.stakingKeeper.GetValidatorByConsAddr(ctx, address)
	if !ok {
		ModuleLogger(ctx).
			Error("can not propagate tompstone: validator not found",
				"validator", address.String())
		e.SlashingKeeper.Tombstone(ctx, address)
		return
	}

	if err := e.k.ScheduleTombstoned(ctx, v.GetOperator()); err != nil {
		ModuleLogger(ctx).
			Error("can not propagate tompstone: scheduler",
				"cause", err,
				"validator", address.String())
	}
	e.SlashingKeeper.Tombstone(ctx, address)
}
