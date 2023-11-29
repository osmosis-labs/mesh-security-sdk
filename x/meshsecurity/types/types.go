package types

type SchedulerTaskType byte

const (
	// SchedulerTaskUndefined null value
	SchedulerTaskUndefined SchedulerTaskType = 0
	// SchedulerTaskHandleEpoch triggered by updates to the virtual staking max cap or by end of epoch
	SchedulerTaskHandleEpoch = 1
	// SchedulerTaskValsetUpdate triggered by any update on the active set. This includes add, remove, validator modifications, slashing, tombstone
	SchedulerTaskValsetUpdate = 2
)
