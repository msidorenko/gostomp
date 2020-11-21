package main

import (
	"github.com/google/uuid"
	"github.com/msidorenko/gostomp"
	"github.com/msidorenko/gostomp/message"
	"os"
	"os/signal"
	"syscall"
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

	msgId := uuid.New().String()
	//We define temporary queue name. Later we will create subscriber for this queue
	tempQueue := "/temp-queue/" + msgId

	//Create a first message
	msg := message.New([]byte("request"))
	msg.SetDestination("/queue/some_queue")
	msg.SetHeader(message.CorrelationId, msgId)
	msg.SetHeader(message.ReplyTo, tempQueue)

	//It's a good idea to create Subscriber for temporary queue before we send message via Producer
	responseSubscription := &gostomp.Subscription{
		Destination: tempQueue,
		Ack:         gostomp.ACK_AUTO,
		Callback: func(msg *message.Message) {
			println("Its our response message: " + string(msg.GetBody()))
		},
	}
	//Create subscription
	err = client.Subscribe(responseSubscription)
	if err != nil {
		println("ERROR: " + err.Error())
	}

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
		Ack:         gostomp.ACK_CLIENT,
		Callback: func(msg *message.Message) {
			//We catch first message.
			body := string(msg.GetBody())

			//And check if we should send reply message back
			responseDestination, err := msg.GetHeader(message.ReplyTo)
			if err != nil {
				//We shouldn't send response message
				println("Error: " + err.Error())
			} else {
				body = body + "-reponse"
				//Create a response message
				responseMessage := message.New([]byte(body))
				responseMessage.SetDestination(responseDestination)
				responseMessage.SetCorrelationId(msg.GetCorrelationId())

				//And sent to back
				err = client.Producer(responseMessage, gostomp.DELIVERY_ASYNC)
				if err != nil {
					//Catch error if Producer cannot send response message.
					println(err.Error())
				}
			}
			client.Ack(msg)
		},
	}

	//Create subscription
	err = client.Subscribe(subscription)
	if err != nil {
		println("ERROR: " + err.Error())
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-sigs:
			err = client.Disconnect()
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}
			println("bye bye :-)")
			os.Exit(0)
			break

		}
	}

}
