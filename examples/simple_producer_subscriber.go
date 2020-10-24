package main

import (
	"github.com/msidorenko/gostomp"
	"github.com/msidorenko/gostomp/message"
	"time"
)

func main() {
	//Create client to Message Broker
	//You can pass login and passcode as http basic auth (tcp://login:passcode@localhost:61613)
	client, err := gostomp.NewClient("tcp://localhost:61613")
	if err != nil {
		panic(err)
	}

	//Create network connection and send auth request
	err = client.Connect()
	if err != nil {
		panic(err)
	}

	//Create new message
	msg := message.New([]byte("Message body"))
	msg.SetDestination("/queue/some_queue")
	msg.SetPersistent(true)
	msg.SetHeader("x-custom-header", time.Now().String())

	//Send message to message broker
	//and wait confirm from server about message accept
	//For async delivery  use gostomp.DELIVERY_ASYNC instead gostomp.DELIVERY_SYNC
	err = client.Producer(msg, gostomp.DELIVERY_SYNC)
	if err != nil {
		panic(err)
	}

	//Create subscription object and define callback function
	//ACK can be: gostomp.ACK_AUTO, gostomp.ACK_CLIENT or gostomp.ACK_CLIENT_INDIVIDUAL
	//Read more about ACK behaviour https://stomp.github.io/stomp-specification-1.2.html#SUBSCRIBE_ack_Header
	subscription := &gostomp.Subscription{
		Destination: "/queue/some_queue",
		Ack:         gostomp.ACK_AUTO,
		Callback: func(msg *message.Message) {
			println("Hey hey, we catch a message: " + string(msg.GetBody()))

		},
	}
	//Create subscription
	err = client.Subscribe(subscription)
	if err != nil {
		println("ERROR: " + err.Error())
	}

	//Wait. Don't forget interrupt process :-)
	for {
	}
}
