package pubsub

import (
	"github.com/x1m3/xevents"
)

var eventBroker *xevents.Handler

func init() {
	eventBroker = xevents.New()
}

func Publish(e Event)  error{
	return eventBroker.Publish(e)
}

func SubscribeFunc(topic string, callback func(event xevents.Event)) {
	eventBroker.RegisterCallback(topic, callback)
}



const UltrahighPriority = 5
const HighPriority = 4
const MedPriority = 3
const LowPriority = 2
const UltraLowPriority = 1

type Event interface {
	xevents.Event
	Error() error
}


