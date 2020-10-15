package frame

import "strings"

const (
	ContentLength = "content-length"
	ContentType   = "content-type"
	Receipt       = "receipt"
	AcceptVersion = "accept-version"
	Host          = "host"
	Version       = "version"
	Login         = "login"
	Passcode      = "passcode"
	HeartBeat     = "heart-beat"
	Session       = "session"
	Server        = "server"
	Destination   = "destination"
	Id            = "id"
	Ack           = "ack"
	Transaction   = "transaction"
	ReceiptId     = "receipt-id"
	Subscription  = "subscription"
	MessageId     = "message-id"
	Message       = "message"
)



type Frame struct {
	Command string
	Headers map[string]string
	Body string
}


func NewFrame(command string, body ...string) *Frame{
	frame := &Frame{
		Command: command,
		Body: strings.Join(body, ""),
		Headers: make(map[string]string),
	}

	return frame
}

func (frame *Frame) AddHeader(key, value string) {
	frame.Headers[key] = value
}
