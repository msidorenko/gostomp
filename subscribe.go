package gostomp

import (
	"github.com/google/uuid"
	"github.com/msidorenko/gostomp/frame"
	"github.com/msidorenko/gostomp/message"
)

const ACK_AUTO = "auto"
const ACK_CLIENT = "client"
const ACK_CLIENT_INDIVIDUAL = "client-individual"

type SubscriptionCallback func(msg *message.Message)

var subscriptions []*Subscription

type Subscription struct {
	id string
	//The valid values for the ack header are auto, client, or client-individual.
	//If the header is not set, it defaults to auto.
	Ack         string
	Destination string
	Callback    SubscriptionCallback
}

func addSubscriptions(subscription *Subscription) {
	subscriptions = append(subscriptions, subscription)
}

func removeSubscription(id string) {

	for i, subscription := range subscriptions {
		if subscription.id == id {
			subscriptions[i] = subscriptions[len(subscriptions)-1] // Copy last element to index i.
			subscriptions[len(subscriptions)-1] = nil              // Erase last element (write zero value).
			subscriptions = subscriptions[:len(subscriptions)-1]   // Truncate slice.
			break
		}
	}
}

func transferFrameToSubscriptions(frm *frame.Frame) {

	for _, subscription := range subscriptions {
		if subscription.id == frm.Headers["subscription"] {
			go subscription.Callback(message.NewFromFrame(frm))
		}
	}
}

func (subs *Subscription) GenerateID() {
	subs.id = uuid.New().String()
}
func (subs *Subscription) GetID() string {
	return subs.id
}
