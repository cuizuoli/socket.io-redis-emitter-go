package enum

const (
	PacketTypeConnect = iota
	PacketTypeDisconnect
	PacketTypeEvent
	PacketTypeAck
	PacketTypeConnectError
	PacketTypeBinaryEvent
	PacketTypeBinaryAck
)
