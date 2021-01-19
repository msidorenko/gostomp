package gostomp

import (
	"crypto/tls"
	"errors"
	"fmt"
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
	Errors     chan error
}

var ReceiptPool = make(map[string]chan *frame.Frame)

//NewClient create object with basic settings for client connection
func NewClient(dsn string) (*Client, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}

	conn := Connection{
		ssl:             false,
		protocol:        u.Scheme,
		addr:            u.Host,
		tryDisconnect:   false,
		heartBeatServer: 0,
		heartBeatClient: 0,
	}

	if u.Scheme == "ssl" {
		conn.protocol = "tcp"
		conn.ssl = true

		insecure := false
		if u.Query().Get("insecure") == "true" {
			insecure = true
		}

		conn.sslConfig = SSLConfig{
			InsecureSkipVerify: insecure,
		}
	}

	if len(u.User.Username()) > 0 {
		conn.login = u.User.Username()
	}

	password, isset := u.User.Password()
	if isset {
		conn.password = password
	}

	hb := u.Query().Get("heart-beat")
	if len(hb) > 0 {
		hbsettings := strings.Split(hb, ",")
		conn.heartBeatClient, err = strconv.ParseInt(hbsettings[0], 10, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(hbsettings[0] + " is not a integer"))
		}
		conn.heartBeatServer, err = strconv.ParseInt(hbsettings[1], 10, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf(hbsettings[1] + " is not a integer"))
		}
	}

	client := &Client{
		connection: conn,
		Errors:     make(chan error),
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

	if client.connection.ssl {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: client.connection.sslConfig.InsecureSkipVerify,
		}

		c := tls.Client(c, tlsConfig)
		err = c.Handshake()
		if err != nil {
			println(err.Error())
		} else {
			client.connection.conn = c
		}
	} else {
		client.connection.conn = c
	}

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

	connectFrame.AddHeader(message.Heartbeat, fmt.Sprint(client.connection.heartBeatClient)+","+fmt.Sprint(client.connection.heartBeatServer))
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

		heartbeat := frm.Headers[message.Heartbeat]
		if len(heartbeat) > 0 {
			hbsettings := strings.Split(heartbeat, ",")

			serverServerHeartBeat, _ := strconv.ParseInt(hbsettings[0], 10, 64)
			if client.connection.heartBeatServer < serverServerHeartBeat {
				client.connection.heartBeatServer = serverServerHeartBeat
			}

			//if hbsettings[1] == 0 then the server does not want to receive heart-beats else client MUST sent msg every max(client.connection.heartBeatClient, hbsettings[1]) milliseconds
			serverClientsHeartBeat, _ := strconv.ParseInt(hbsettings[1], 10, 64)
			if client.connection.heartBeatClient < serverClientsHeartBeat {
				client.connection.heartBeatClient = serverClientsHeartBeat
			}
		}

		client.session = make([]Session, 0)
		client.session = append(client.session, Session{id: frm.Headers[message.Session]})
	}

	//Start gourtine for continuously read from socket
	go client.readerLoop(reader)
	return nil
}

//Disconnect method will close connection with Message broker
//Client send DISCONNECT header with receipt header and wait for ack from message broker
//@TODO add support multiply servers
func (client *Client) Disconnect() error {

	client.connection.tryDisconnect = true

	frm := frame.NewFrame(frame.DISCONNECT, nil)
	frm.Headers[message.Receipt] = "77"
	ReceiptPool["77"] = make(chan *frame.Frame)

	err := client.sender(frm)
	if err != nil {
		return err
	}

	_, ok := <-ReceiptPool["77"]
	if !ok {
		return errors.New("ERROR: cannot get receipt  frame from channel")
	}
	close(ReceiptPool["77"])
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

	if frm.Command != frame.DISCONNECT && client.connection.tryDisconnect {
		return errors.New("Disconnect in progress. Clients MUST NOT send any more frames after the DISCONNECT frame is sent.")
	}

	writer := NewWriter(client.connection.conn, 4096)
	err := writer.Write(frm)
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) readerLoop(reader *Reader) {

	for {
		frm, err := reader.Read()
		if err != nil {
			client.Errors <- err
			return
		}
		if frm == nil {
			//heart-beat
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
