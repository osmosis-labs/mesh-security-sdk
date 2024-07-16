package types

import (
	errorsmod "cosmossdk.io/errors"
	"encoding/binary"
)

func (c ConsumerPacketData) GetBytes() []byte {
	var bytes []byte
	bytes = append(bytes, Int32ToBytes(int32(c.Type))...)
	if c.Type != PipedValsetOperation_VALIDATOR_SLASHED {
		packetScheduleInfo := c.GetSchedulePacketData()
		bytes = append(bytes, ModuleCdc.MustMarshalJSON(packetScheduleInfo)...)
	} else {
		packetSlashInfo := c.GetSlashPacketData()
		bytes = append(bytes, ModuleCdc.MustMarshalJSON(packetSlashInfo)...)
	}
	return bytes
}

func UnmarshalConsumerPacketData(data []byte) (ConsumerPacketData, error) {
	cp := PipedValsetOperation(BigEndianToInt32(data[:4]))
	if cp != PipedValsetOperation_VALIDATOR_SLASHED {
		var s ScheduleInfo
		err := ModuleCdc.UnmarshalJSON(data[4:], &s)
		if err != nil {
			return ConsumerPacketData{}, err
		}
		return ConsumerPacketData{
			Type: cp,
			Data: &ConsumerPacketData_SchedulePacketData{
				SchedulePacketData: &s,
			},
		}, nil
	} else {
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

type PacketAckResult []byte

var (
	// Slash packet handled result ack, sent by a throttling provider to indicate that a slash packet was handled.
	SlashPacketHandledResult = PacketAckResult([]byte{byte(1)})
	// Slash packet bounced result ack, sent by a throttling provider to indicate that a slash packet was NOT handled
	// and should eventually be retried.
	SlashPacketBouncedResult = PacketAckResult([]byte{byte(2)})
)

func NewConsumerPacketData(cpdType PipedValsetOperation, data isConsumerPacketData_Data) ConsumerPacketData {
	return ConsumerPacketData{
		Type: cpdType,
		Data: data,
	}
}

func (s SlashInfo) Validate() error {
	if s.Validator == "" {
		return errorsmod.Wrap(ErrInvalidPacketData, "invalid validator")
	}
	if s.Power == 0 {
		return errorsmod.Wrap(ErrInvalidPacketData, "validator power cannot be zero")
	}
	if s.InfractionHeight <= 0 {
		return errorsmod.Wrap(ErrInvalidPacketData, "Infraction Height cannot be zero")
	}
	if s.SlashFraction == "0" {
		return errorsmod.Wrap(ErrInvalidPacketData, "Slash Fraction cannot be zero")
	}
	if s.TotalSlashAmount == "0" {
		return errorsmod.Wrap(ErrInvalidPacketData, "Total Slash Amount cannot be zero")
	}
	if s.TimeInfraction < 0 {
		return errorsmod.Wrap(ErrInvalidPacketData, "TimeInfraction cannot be negative")
	}
	return nil
}
