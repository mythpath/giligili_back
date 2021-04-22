package pubsub

import (
	"encoding/json"
)

type Event struct {
	broker  *Broker
	id      uint
	topic   string
	message string
}

func (e *Event) GetEventContent(v interface{}) error {
	return json.Unmarshal([]byte(e.message), v)
}

func (e *Event) Ack() error {
	return e.broker.deleteTopicmsg(e.id)
}
