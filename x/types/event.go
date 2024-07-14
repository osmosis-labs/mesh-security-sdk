package types

const (
	EventTypeTimeout            = "timeout"
	EventTypePacket             = "mesh_packet"
	EventTypeChannelEstablished = "channel_established"

	AttributeKeyAckSuccess = "success"
	AttributeKeyAck        = "acknowledgement"
	AttributeKeyAckError   = "error"
	AttributeChainID       = "chain_id"
)
