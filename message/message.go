package message

import (
	"errors"
	"github.com/google/uuid"
	"github.com/msidorenko/gostomp/frame"
	"strconv"
	"time"
)

const (
	Host      = "host"
	Version   = "version"
	Login     = "login"
	Passcode  = "passcode"
	Heartbeat = "heart-beat"

	ContentLength = "content-length"
	ContentType   = "content-type"
	AcceptVersion = "accept-version"
	Session       = "session"
	Server        = "server"
	Destination   = "destination"
	Id            = "id"
	Ack           = "ack"
	ReceiptId     = "receipt-id"
	Subscription  = "subscription"
	MessageId     = "message-id"
	Message_      = "message"
	Receipt       = "receipt"
	TimeStamp     = "timestamp "

	//ActiveMQ specific
	ReplyTo       = "reply-to"
	CorrelationId = "correlation-id"
	Persistent    = "persistent"
	Delay         = "AMQ_SCHEDULED_DELAY"
)

type Message struct {
	headers map[string]string
	body    []byte
}

func NewFromFrame(frm *frame.Frame) *Message {
	msg := &Message{
		body:    frm.Body,
		headers: frm.Headers,
	}
	return msg
}

func New(body []byte) *Message {
	msg := &Message{
		body:    body,
		headers: make(map[string]string),
	}

	msg.SetHeader(MessageId, uuid.New().String())

	return msg
}

func (message *Message) SetBody(body []byte) {
	message.body = body
}

func (message *Message) GetBody() []byte {
	return message.body
}

func (message *Message) SetHeader(key, value string) {
	message.headers[key] = value
}

func (message *Message) GetHeaders() map[string]string {
	return message.headers
}

func (message *Message) GetHeader(key string) (string, error) {
	if value, ok := message.headers[key]; ok {
		return value, nil
	} else {
		return "", errors.New("Header '" + key + "' did not exist")
	}
}

func (message *Message) SetID(id string) {
	message.headers[MessageId] = id
}

func (message *Message) GetID() string {

	id, err := message.GetHeader(MessageId)
	if err != nil {
		return ""
	}
	return id
}

func (message *Message) SetPersistent(persistence bool) {
	if persistence {
		message.headers[Persistent] = "true"
	} else {
		message.headers[Persistent] = "false"
	}
}

func (message *Message) GetPersistent() bool {

	if value, ok := message.headers[Persistent]; !ok {
		//STOMP messages are non-persistent by default.
		return false
	} else {
		if value == "true" {
			return true
		} else {
			return false
		}
	}
}

func (message *Message) SetDestination(destination string) {
	message.headers[Destination] = destination
}

func (message *Message) GetDestination() string {
	return message.headers[Destination]
}

func (message *Message) SetDelay(delay time.Duration) {
	message.headers[Delay] = strconv.FormatInt(delay.Milliseconds(), 10)
}

func (message *Message) SetCorrelationId(correlationId string) {
	message.headers[CorrelationId] = correlationId
}

func (message *Message) GetCorrelationId() string {
	if value, ok := message.headers[CorrelationId]; !ok {
		return ""
	} else {
		return value
	}

}
