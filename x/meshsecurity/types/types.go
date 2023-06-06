package types

type SchedulerTaskType byte

const (
	SchedulerTaskUndefined SchedulerTaskType = 0
	SchedulerTaskRebalance                   = 1
)
