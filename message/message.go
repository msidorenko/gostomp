package message

import (
	"errors"
	"github.com/google/uuid"
	"github.com/msidorenko/gostomp/frame"
)

type Message struct {
	headers map[string]string
	body string
}

func NewFromFrame(frm *frame.Frame) *Message {
	msg := &Message{
		body: frm.Body,
		headers: frm.Headers,
	}
	return msg
}

func New(body string) *Message {
	msg := &Message{
		body: body,
		headers: make(map[string]string),
	}

	msg.SetHeader(frame.MessageId, uuid.New().String())

	return msg
}

func (message *Message) SetBody(body string){
	message.body = body
}

func (message *Message) GetBody() string {
	return message.body
}

func (message *Message) SetHeader(key, value string){
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

func (message *Message) SetID(id string){
	message.headers[frame.MessageId] = id
}

func (message *Message) GetID() string {

	id, err := message.GetHeader(frame.MessageId)
	if err != nil {
		return ""
	}
	return id
}

func (message *Message) SetDestination(destination string) {
	message.headers["destination"] = destination
}

func (message *Message) GetDestination() string {
	return message.headers["destination"]
}