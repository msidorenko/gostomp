package frame

const (
	// Connect commands.
	CONNECT   = "CONNECT"
	STOMP     = "STOMP"
	CONNECTED = "CONNECTED"

	// Client commands.
	SEND        = "SEND"
	SUBSCRIBE   = "SUBSCRIBE"
	UNSUBSCRIBE = "UNSUBSCRIBE"
	ACK         = "ACK"
	NACK        = "NACK"
	BEGIN       = "BEGIN"
	COMMIT      = "COMMIT"
	ABORT       = "ABORT"
	DISCONNECT  = "DISCONNECT"

	// Server commands.
	MESSAGE = "MESSAGE"
	RECEIPT = "RECEIPT"
	ERROR   = "ERROR"
)

