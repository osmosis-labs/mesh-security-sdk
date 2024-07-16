package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBytesAndGetData(t *testing.T) {
	c1 := ConsumerPacketData{
		Type: PipedValsetOperation_VALIDATOR_SLASHED,
		Data: &ConsumerPacketData_SlashPacketData{
			SlashPacketData: &SlashInfo{
				InfractionHeight: 1,
				Power:            2,
				TotalSlashAmount: "3",
				SlashFraction:    "0.1",
				TimeInfraction:   1000,
			},
		},
	}

	data := c1.GetBytes()
	c2, err := UnmarshalConsumerPacketData(data)
	require.NoError(t, err)
	require.Equal(t, c1, c2)

	require.Equal(t, int64(1), c2.GetSlashPacketData().InfractionHeight)
	require.Equal(t, int64(2), c2.GetSlashPacketData().Power)
	require.Equal(t, "3", c2.GetSlashPacketData().TotalSlashAmount)
	require.Equal(t, "0.1", c2.GetSlashPacketData().SlashFraction)
	require.Equal(t, int64(1000), c2.GetSlashPacketData().TimeInfraction)

	s1 := ConsumerPacketData{
		Type: PipedValsetOperation_VALIDATOR_TOMBSTONED,
		Data: &ConsumerPacketData_SchedulePacketData{
			SchedulePacketData: &ScheduleInfo{
				Validator: "val",
				Actor:     "contract",
			},
		},
	}
	data = s1.GetBytes()
	s2, err := UnmarshalConsumerPacketData(data)
	require.NoError(t, err)
	require.Equal(t, s1, s2)
}
