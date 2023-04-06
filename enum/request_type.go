package enum

/**
 * Request types, for messages between nodes
 */
const (
	RequestTypeSockets = iota
	RequestTypeAllRooms
	RequestTypeRemoteJoin
	RequestTypeRemoteLeave
	RequestTypeRemoteDisconnect
	RequestTypeRemoteFetch
	RequestTypeServerSideEmit
)
