package types

import (
	"encoding/binary"
)

func (c ConsumerPacketData) GetBytes() []byte {
	var bytes []byte
	bytes = append(bytes, Int32ToBytes(int32(c.Type))...)
	if c.Type != PipedValsetOperation_VALIDATOR_SLASHED {
		bytes = append(bytes, ModuleCdc.MustMarshalJSON(&c)...)
		return bytes
	}

	packetSlashInfo := c.GetSlashPacketData()
	bytes = append(bytes, ModuleCdc.MustMarshalJSON(packetSlashInfo)...)
	return bytes
}

func UnmarshalConsumerPacketData(data []byte) (ConsumerPacketData, error) {
	cp := PipedValsetOperation(BigEndianToInt32(data[:4]))
	if cp != PipedValsetOperation_VALIDATOR_SLASHED {
		var c ConsumerPacketData
		ModuleCdc.UnmarshalJSON(data[4:], &c)
		return c, nil
	}

	var slash SlashInfo
	err := ModuleCdc.UnmarshalJSON(data[4:], &slash)
	if err != nil {
		return ConsumerPacketData{}, err
	}

	return ConsumerPacketData{
		Type: cp,
		Data: &ConsumerPacketData_SlashPacketData{
			SlashPacketData: &slash,
		},
	}, nil

}

func BigEndianToInt32(bz []byte) int32 {
	if len(bz) == 0 {
		return 0
	}
	return int32(binary.BigEndian.Uint32(bz))
}

func Int32ToBytes(i int32) []byte {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(i))
	return data
}
