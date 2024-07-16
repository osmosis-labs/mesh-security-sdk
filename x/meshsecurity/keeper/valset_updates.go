package keeper

import (
	"cosmossdk.io/math"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	cptypes "github.com/osmosis-labs/mesh-security-sdk/x/types"

	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/contract"
	"github.com/osmosis-labs/mesh-security-sdk/x/meshsecurity/types"
)

// ScheduleBonded store a validator update to bonded status for the valset update report
func (k Keeper) ScheduleBonded(ctx sdk.Context, addr sdk.ValAddress) error {
	packet := cptypes.ConsumerPacketData{
		Type: cptypes.PipedValsetOperation_VALIDATOR_BONDED,
		Data: &cptypes.ConsumerPacketData_SchedulePacketData{
			SchedulePacketData: &cptypes.ScheduleInfo{
				Validator: addr.String(),
			},
		},
	}
	return k.sendAsync(ctx, cptypes.PipedValsetOperation_VALIDATOR_BONDED, addr, packet)
}

// ScheduleUnbonded store a validator update to unbonded status for the valset update report
func (k Keeper) ScheduleUnbonded(ctx sdk.Context, addr sdk.ValAddress) error {
	packet := cptypes.ConsumerPacketData{
		Type: cptypes.PipedValsetOperation_VALIDATOR_UNBONDED,
		Data: &cptypes.ConsumerPacketData_SchedulePacketData{
			SchedulePacketData: &cptypes.ScheduleInfo{
				Validator: addr.String(),
			},
		},
	}
	return k.sendAsync(ctx, cptypes.PipedValsetOperation_VALIDATOR_UNBONDED, addr, packet)
}

// ScheduleSlashed store a validator slash event / data for the valset update report
func (k Keeper) ScheduleSlashed(ctx sdk.Context, valAddr sdk.ValAddress, slashRatio sdk.Dec) error {
	power := k.Staking.GetLastValidatorPower(ctx, valAddr)
	infractionHeight := ctx.BlockHeight() - sdk.ValidatorUpdateDelay - 1
	validator, _ := k.Staking.GetValidator(ctx, valAddr)
	totalSlashAmount := sdk.NewDecFromInt(validator.Tokens).MulTruncate(slashRatio).RoundInt()
	// TODO: timeInfraction

	var slashInfo = &cptypes.SlashInfo{
		Validator:        validator.OperatorAddress,
		Power:            power,
		InfractionHeight: infractionHeight,
		TotalSlashAmount: totalSlashAmount.String(),
		SlashFraction:    slashRatio.String(),
		// TODO: timeInfraction
		TimeInfraction: 0,
	}

	packet := cptypes.ConsumerPacketData{
		Type: cptypes.PipedValsetOperation_VALIDATOR_SLASHED,
		Data: &cptypes.ConsumerPacketData_SlashPacketData{
			SlashPacketData: slashInfo,
		},
	}
	return k.sendAsync(ctx, cptypes.PipedValsetOperation_VALIDATOR_SLASHED, valAddr, packet)
}

// ScheduleJailed store a validator update to jailed status for the valset update report
func (k Keeper) ScheduleJailed(ctx sdk.Context, addr sdk.ValAddress) error {
	packet := cptypes.ConsumerPacketData{
		Type: cptypes.PipedValsetOperation_VALIDATOR_JAILED,
		Data: &cptypes.ConsumerPacketData_SchedulePacketData{
			SchedulePacketData: &cptypes.ScheduleInfo{
				Validator: addr.String(),
			},
		},
	}
	return k.sendAsync(ctx, cptypes.PipedValsetOperation_VALIDATOR_JAILED, addr, packet)
}

// ScheduleTombstoned store a validator update to tombstoned status for the valset update report
func (k Keeper) ScheduleTombstoned(ctx sdk.Context, addr sdk.ValAddress) error {
	packet := cptypes.ConsumerPacketData{
		Type: cptypes.PipedValsetOperation_VALIDATOR_TOMBSTONED,
		Data: &cptypes.ConsumerPacketData_SchedulePacketData{
			SchedulePacketData: &cptypes.ScheduleInfo{
				Validator: addr.String(),
			},
		},
	}
	return k.sendAsync(ctx, cptypes.PipedValsetOperation_VALIDATOR_TOMBSTONED, addr, packet)
}

// ScheduleUnjailed store a validator update to unjailed status for the valset update report
func (k Keeper) ScheduleUnjailed(ctx sdk.Context, addr sdk.ValAddress) error {
	packet := cptypes.ConsumerPacketData{
		Type: cptypes.PipedValsetOperation_VALIDATOR_UNJAILED,
		Data: &cptypes.ConsumerPacketData_SchedulePacketData{
			SchedulePacketData: &cptypes.ScheduleInfo{
				Validator: addr.String(),
			},
		},
	}
	return k.sendAsync(ctx, cptypes.PipedValsetOperation_VALIDATOR_UNJAILED, addr, packet)
}

// ScheduleModified store a validator metadata update for the valset update report
func (k Keeper) ScheduleModified(ctx sdk.Context, addr sdk.ValAddress) error {
	packet := cptypes.ConsumerPacketData{
		Type: cptypes.PipedValsetOperation_VALIDATOR_MODIFIED,
		Data: &cptypes.ConsumerPacketData_SchedulePacketData{
			SchedulePacketData: &cptypes.ScheduleInfo{
				Validator: addr.String(),
			},
		},
	}
	return k.sendAsync(ctx, cptypes.PipedValsetOperation_VALIDATOR_MODIFIED, addr, packet)
}

// instead of sync calls to the contracts for the different kind of valset changes in a block, we store them in the mem db
// and async send to all registered contracts in the end blocker
func (k Keeper) sendAsync(ctx sdk.Context, op cptypes.PipedValsetOperation, valAddr sdk.ValAddress, packet cptypes.ConsumerPacketData) error {
	// if op
	ModuleLogger(ctx).Debug("storing for async update", "operation", int(op), "val", valAddr.String())
	value := packet.GetBytes()
	ctx.KVStore(k.memKey).Set(types.BuildPipedValsetOpKey(op, valAddr), value)
	// and schedule an update callback for all registered contracts
	var innerErr error
	k.IterateMaxCapLimit(ctx, func(contractAddr sdk.AccAddress, m math.Int) bool {
		if m.GT(math.ZeroInt()) {
			innerErr = k.ScheduleOneShotTask(ctx, types.SchedulerTaskValsetUpdate, contractAddr, uint64(ctx.BlockHeight()))
			if innerErr != nil {
				return true
			}
		}
		return false
	})
	return innerErr
}

// ValsetUpdateReport aggregate all stored changes of the current block. Should be called by an end-blocker.
// The events reported are categorized by type and not time. Conflicting events as Bonded/ Unbonded
// are not supposed to happen within the same block
func (k Keeper) ValsetUpdateReport(ctx sdk.Context) (contract.ValsetUpdate, error) {
	var innerErr error
	appendValidator := func(set *[]wasmvmtypes.Validator, valAddr sdk.ValAddress) bool {
		val, ok := k.Staking.GetValidator(ctx, valAddr)
		if !ok {
			innerErr = types.ErrUnknown.Wrapf("validator %s", valAddr)
			return true
		}
		*set = append(*set, ConvertSdkValidatorToWasm(val))
		return false
	}
	slashValidator := func(set *[]contract.ValidatorSlash, valAddr sdk.ValAddress, power int64, infractionHeight int64,
		infractionTime int64, slashAmount string, slashRatio string) bool {
		valSlash := contract.ValidatorSlash{
			ValidatorAddr:    valAddr.String(),
			Power:            power,
			InfractionHeight: infractionHeight,
			InfractionTime:   infractionTime,
			Height:           ctx.BlockHeight(),
			Time:             ctx.BlockTime().Unix(),
			SlashAmount:      slashAmount,
			SlashRatio:       slashRatio,
		}
		*set = append(*set, valSlash)
		return false
	}
	r := contract.ValsetUpdate{ // init with empty slices for contract that does not handle null or omitted fields
		Additions:  make([]contract.Validator, 0),
		Removals:   make([]contract.ValidatorAddr, 0),
		Updated:    make([]contract.Validator, 0),
		Jailed:     make([]contract.ValidatorAddr, 0),
		Unjailed:   make([]contract.ValidatorAddr, 0),
		Tombstoned: make([]contract.ValidatorAddr, 0),
		Slashed:    make([]contract.ValidatorSlash, 0),
	}
	err := k.iteratePipedValsetOperations(ctx, func(packet *cptypes.ConsumerPacketData) bool {
		switch packet.Type {
		case cptypes.PipedValsetOperation_VALIDATOR_BONDED:
			data := packet.GetSchedulePacketData()
			return appendValidator(&r.Additions, sdk.ValAddress(data.Validator))
		case cptypes.PipedValsetOperation_VALIDATOR_UNBONDED:
			data := packet.GetSchedulePacketData()
			r.Removals = append(r.Removals, data.Validator)
		case cptypes.PipedValsetOperation_VALIDATOR_JAILED:
			data := packet.GetSchedulePacketData()
			r.Jailed = append(r.Jailed, data.Validator)
		case cptypes.PipedValsetOperation_VALIDATOR_TOMBSTONED:
			data := packet.GetSchedulePacketData()
			r.Tombstoned = append(r.Tombstoned, data.Validator)
		case cptypes.PipedValsetOperation_VALIDATOR_UNJAILED:
			data := packet.GetSchedulePacketData()
			r.Unjailed = append(r.Unjailed, data.Validator)
		case cptypes.PipedValsetOperation_VALIDATOR_MODIFIED:
			data := packet.GetSchedulePacketData()
			return appendValidator(&r.Updated, sdk.ValAddress(data.Validator))
		case cptypes.PipedValsetOperation_VALIDATOR_SLASHED:
			data := packet.GetSlashPacketData()
			return slashValidator(&r.Slashed, sdk.ValAddress(data.Validator), data.Power, data.InfractionHeight, data.TimeInfraction,
				data.TotalSlashAmount, data.SlashFraction)
		default:
			innerErr = types.ErrInvalid.Wrapf("undefined operation type %X", packet.Type)
			return true
		}
		return false
	})
	if err != nil {
		return r, err
	}
	return r, innerErr
}

// ClearPipedValsetOperations delete all entries from the temporary store that contains the valset updates.
func (k Keeper) ClearPipedValsetOperations(ctx sdk.Context) {
	var keys [][]byte
	pStore := prefix.NewStore(ctx.KVStore(k.memKey), types.PipedValsetPrefix)
	iter := pStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}
	_ = iter.Close()
	for _, k := range keys {
		pStore.Delete(k)
	}
}

// iterate through all stored valset updates. Due to the storage key, there are no contract duplicates within an operation type.
func (k Keeper) iteratePipedValsetOperations(ctx sdk.Context, cb func(packet *cptypes.ConsumerPacketData) bool) error {
	pStore := prefix.NewStore(ctx.KVStore(k.memKey), types.PipedValsetPrefix)
	iter := pStore.Iterator(nil, nil)
	for ; iter.Valid(); iter.Next() {
		consumerPacket, err := cptypes.UnmarshalConsumerPacketData(iter.Value())
		if err != nil {
			panic(err)
		}
		if cb(&consumerPacket) {
			break
		}
	}
	return iter.Close()
}

// ConvertSdkValidatorToWasm helper method
func ConvertSdkValidatorToWasm(v stakingtypes.Validator) wasmvmtypes.Validator {
	return wasmvmtypes.Validator{
		Address:       v.OperatorAddress,
		Commission:    v.Commission.Rate.String(),
		MaxCommission: v.Commission.MaxRate.String(),
		MaxChangeRate: v.Commission.MaxChangeRate.String(),
	}
}
