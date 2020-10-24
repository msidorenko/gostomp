package frame

const (
	// Connections commands
	CONNECT    = "CONNECT"
	CONNECTED  = "CONNECTED"
	STOMP      = "STOMP"
	DISCONNECT = "DISCONNECT"

	// Client commands
	SEND        = "SEND"
	SUBSCRIBE   = "SUBSCRIBE"
	UNSUBSCRIBE = "UNSUBSCRIBE"
	ACK         = "ACK"
	NACK        = "NACK"

	// Server commands
	MESSAGE = "MESSAGE"
	RECEIPT = "RECEIPT"
	ERROR   = "ERROR"
)
