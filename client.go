package gostomp

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

//NewClient create object with basic settings for client connection
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

//Connect method establish connection with the Message Broker server and authorize via CONNECT command
//@TODO check all of servers for failover
func (client *Client) Connect() error {
	c, err := net.Dial(client.connection.protocol, client.connection.addr)
	if err != nil {
		return err
	}

	client.connection.conn = c
	//After established network connection, we try send CONNECT frame to the message broker
	connectFrame := frame.NewFrame(frame.CONNECT, []byte(""))

	//In some cases we need to do auth by login and password
	if client.connection.login != "" {
		connectFrame.AddHeader(message.Login, client.connection.login)
		if client.connection.password != "" {
			connectFrame.AddHeader(message.Passcode, client.connection.password)
		}
	}

	connectFrame.AddHeader(message.AcceptVersion, "1.0,1.1,1.2")

	host, _, err := net.SplitHostPort(client.connection.addr)
	if err == nil {
		connectFrame.AddHeader(message.Host, host)
	}

	connectFrame.AddHeader(message.Receipt, uuid.New().String())

	err = client.sender(connectFrame)
	if err != nil {
		return err
	}

	reader := NewReader(client.connection.conn, 4096)
	frm, err := reader.Read()
	if err != nil {
		return err
	}

	if frm != nil {
		client.connection.server = frm.Headers[message.Server]
		client.connection.version = strings.Split(frm.Headers[message.Session], ",")
		client.connection.heaetbeat = frm.Headers[message.Heartbeat]
		client.session = make([]Session, 0)
		client.session = append(client.session, Session{id: frm.Headers[message.Session]})
	}

	//Start gourtine for continuously read from socket
	go readLoop(reader)
	return nil
}

//Producer method send a Message to the Message Broker
//msg *message.Message
//deliveryMode bool
//async - just push frame to the socket and forget about it. deliveryMode == false
//sync - push frame to the socket and wait confirm message from the Message broker. deliveryMode == true
func (client *Client) Producer(msg *message.Message, deliveryMode bool) error {
	frm := frame.NewFrame("SEND", msg.GetBody())
	frm.Headers[message.Destination] = msg.GetDestination()
	frm.Headers[message.ContentLength] = strconv.Itoa(len(msg.GetBody()))
	frm.Headers[message.ContentType] = "text/plain"

	for k, v := range msg.GetHeaders() {
		frm.Headers[k] = v
	}

	msgId := msg.GetID()
	if msgId != "" {
		frm.Headers[message.MessageId] = msgId
	} else {
		msgId := uuid.New().String()
		msg.SetID(msgId)
		frm.Headers[message.MessageId] = msgId
	}

	if deliveryMode == DELIVERY_SYNC {
		frm.Headers[message.Receipt] = msgId
		ReceiptPool[msgId] = make(chan *frame.Frame)
	}

	err := client.sender(frm)
	if err != nil {
		return err
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

//Subscribe send SUBSCRIBE command to the Message Broker
func (client *Client) Subscribe(subscription *Subscription) error {
	subscription.GenerateID()

	frm := frame.NewFrame(frame.SUBSCRIBE, []byte(""))
	frm.Headers[message.Destination] = subscription.Destination
	frm.Headers[message.Id] = subscription.GetID()

	if subscription.Ack == "" {
		frm.Headers[message.Ack] = ACK_AUTO
	} else {
		frm.Headers[message.Ack] = subscription.Ack
	}

	err := client.sender(frm)
	if err != nil {
		println("Error: " + err.Error())
		return errors.New("Cannot subscribe to " + subscription.Destination + ". Reason: " + err.Error())
	}

	addSubscriptions(subscription)
	return nil
}

func (client *Client) Unsubscribe(subscriptionId string) {
	frm := frame.NewFrame(frame.UNSUBSCRIBE, []byte(""))
	frm.Headers[message.Id] = subscriptionId

	err := client.sender(frm)
	if err != nil {
		println("Error: " + err.Error())
	}

	removeSubscription(subscriptionId)
}

func (client *Client) Ack(msg *message.Message) {
	frm := frame.NewFrame(frame.ACK, []byte(""))
	ackId, err := msg.GetHeader(message.Ack)
	if err != nil {
		return
	} else {
		frm.Headers[message.Id] = ackId
	}

	err = client.sender(frm)
	if err != nil {
		println("Error: " + err.Error())
	}

}

func (client *Client) NAck(msg *message.Message) {
	frm := frame.NewFrame(frame.NACK, []byte(""))
	ackId, err := msg.GetHeader(message.Ack)
	if err != nil {
		return
	} else {
		frm.Headers[message.Id] = ackId
	}

	err = client.sender(frm)
	if err != nil {
		println("Error: " + err.Error())
	}
}

func (client *Client) sender(frm *frame.Frame) error {
	writer := NewWriter(client.connection.conn, 4096)
	err := writer.Write(frm)
	if err != nil {
		return err
	}
	return nil
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
			ReceiptPool[frm.Headers[message.ReceiptId]] <- frm
			break
		case frame.ERROR:
			break
		}

	}
}
