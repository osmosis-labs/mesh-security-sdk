package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBytesAndX(t *testing.T) {
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
}
