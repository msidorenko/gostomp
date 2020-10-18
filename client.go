package stomp

import (
	"errors"
	"github.com/google/uuid"
	"github.com/msidorenko/gostomp/frame"
	"github.com/msidorenko/gostomp/message"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const DELIVERY_SYNC = true
const DELIVERY_ASYNC = false

type Client struct {
	connection Connection
	session    []Session
}

var ReceiptPool = make(map[string]chan *frame.Frame)

//NewClient creat struct with basic settings for client connection
func NewClient(dsn string) (*Client, error) {

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	conn := Connection{
		protocol: u.Scheme,
		addr:     u.Host,
	}

	if len(u.User.Username()) > 0 {
		conn.login = u.User.Username()
	}

	password, isset := u.User.Password()
	if isset {
		conn.password = password
	}

	client := &Client{
		connection: conn,
	}
	return client, nil
}

func (client *Client) Connect() error {

	c, err := net.Dial(client.connection.protocol, client.connection.addr)
	if err != nil {
		return err
	}

	client.connection.conn = c

	connectFrame := frame.NewFrame(frame.CONNECT, "")
	if client.connection.login != "" {
		connectFrame.AddHeader(frame.Login, client.connection.login)
		if client.connection.password != "" {
			connectFrame.AddHeader(frame.Passcode, client.connection.password)
		}
	}

	connectFrame.AddHeader(frame.AcceptVersion, "1.0,1.1,1.2")

	host, _, err := net.SplitHostPort(client.connection.addr)
	if err == nil {
		connectFrame.AddHeader(frame.Host, host)
	}

	connectFrame.AddHeader(frame.Receipt, uuid.New().String())

	writer := NewWriter(client.connection.conn)
	reader := NewReader(client.connection.conn)

	err = writer.Write(connectFrame)
	if err != nil {
		return err
	}

	frm, err := reader.Read()
	if err != nil {
		return err
	}

	if frm != nil {
		client.connection.server = frm.Headers[frame.Server]
		client.connection.version = strings.Split(frm.Headers[frame.Session], ",")
		client.connection.heaetbeat = frm.Headers[frame.HeartBeat]
		client.session = make([]Session, 0)
		client.session = append(client.session, Session{id: frm.Headers[frame.Session]})
	}

	go readLoop(reader)
	return nil
}

func (client *Client) Producer(msg *message.Message, deliveryMode bool) error {
	frm := frame.NewFrame("SEND", msg.GetBody())
	frm.Headers[frame.Destination] = msg.GetDestination()
	frm.Headers[frame.ContentLength] = strconv.Itoa(len(msg.GetBody()))
	frm.Headers[frame.ContentType] = "text/plain"

	for k, v := range msg.GetHeaders() {
		frm.Headers[k] = v
	}

	msgId := msg.GetID()
	if msgId != "" {
		frm.Headers[frame.MessageId] = msgId
	} else {
		msgId := uuid.New().String()
		msg.SetID(msgId)
		frm.Headers[frame.MessageId] = msgId
	}

	if deliveryMode == DELIVERY_SYNC {
		frm.Headers[frame.Receipt] = msgId
		ReceiptPool[msgId] = make(chan *frame.Frame)
	}

	writer := NewWriter(client.connection.conn)

	err := writer.Write(frm)
	if err != nil {
		println("Error: " + err.Error())
	}

	if deliveryMode == DELIVERY_SYNC {
		_, ok := <-ReceiptPool[msgId]
		if !ok {
			return errors.New("ERROR: cannot get receipt  frame from channel")
		}
		close(ReceiptPool[msgId])
	}
	return nil
}

func (client *Client) Subscribe(subscription *Subscription) error {

	subscription.GenerateID()

	frm := frame.NewFrame(frame.SUBSCRIBE, "")
	frm.Headers[frame.Destination] = subscription.Destination
	frm.Headers[frame.Id] = subscription.GetID()

	if subscription.Ack == "" {
		frm.Headers[frame.Ack] = ACK_AUTO
	} else {
		frm.Headers[frame.Ack] = subscription.Ack
	}

	writer := NewWriter(client.connection.conn)
	err := writer.Write(frm)
	if err != nil {
		println("Error: " + err.Error())
		return errors.New("Cannot subscribe to " + subscription.Destination + ". Reason: " + err.Error())
	}

	addSubscriptions(subscription)
	return nil
}

func (client *Client) Unsubscribe(subscriptionId string) {

	frm := frame.NewFrame(frame.UNSUBSCRIBE)
	frm.Headers[frame.Id] = subscriptionId

	writer := NewWriter(client.connection.conn)
	err := writer.Write(frm)
	if err != nil {
		println("Error: " + err.Error())
	}

	removeSubscription(subscriptionId)

}

func (client *Client) Ack(message *message.Message) {
	frm := frame.NewFrame(frame.ACK, "")
	ackId, err := message.GetHeader(frame.Ack)
	if err != nil {
		return
	} else {
		frm.Headers[frame.Id] = ackId
	}

	writer := NewWriter(client.connection.conn)
	err = writer.Write(frm)
	if err != nil {
		println("Error: " + err.Error())
	}

}

func (client *Client) NAck(message *message.Message) {
	frm := frame.NewFrame(frame.NACK, "")
	ackId, err := message.GetHeader(frame.Ack)
	if err != nil {
		return
	} else {
		frm.Headers[frame.Id] = ackId
	}

	writer := NewWriter(client.connection.conn)
	err = writer.Write(frm)
	if err != nil {
		println("Error: " + err.Error())
	}

}

func readLoop(reader *Reader) {

	for {
		frm, err := reader.Read()
		if err != nil {
			return
		}
		if frm == nil {
			continue
		}
		switch frm.Command {
		case frame.MESSAGE:
			transferFrameToSubscriptions(frm)
			break
		case frame.RECEIPT:
			ReceiptPool[frm.Headers[frame.ReceiptId]] <- frm
			break
		}

	}
}
