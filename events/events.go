package events

import (
	"encoding/json"
	"errors"
	"hash/fnv"
	"net"
	"runtime"
	"trikmq/channels"
)

type EventType int

const (
	SUBSCRIBE EventType = iota
	UNSUBSCRIBE
	BROADCAST
)

type EventHeader struct {
	Type EventType `json:"type"`
}

type Event struct {
	Headers *EventHeader `json:"header"`
	Payload interface{}  `json:"payload"`
}

type EventManager struct {
	channels map[string]*channels.Channel
}

func NewEventManager() *EventManager {
	return &EventManager{
		channels: make(map[string]*channels.Channel),
	}
}

func hash(channelName string) uint64 {
	h := fnv.New64()
	h.Write([]byte(channelName))
	return h.Sum64()
}

func (e *EventManager) registerChannel(channelName string) uint64 {
	secret := hash(channelName)
	e.channels[channelName] = channels.NewChannel(secret)
	return secret
}

func (e *EventManager) ParseEvent(msg *string) (*Event, error) {
	var event Event
	err := json.Unmarshal([]byte(*msg), &event)

	if err != nil {
		return nil, errors.New("failed to parse message")
	}
	return &event, nil
}

func (e *EventManager) HandleEvent(conn net.Conn, event *Event) {
	switch event.Headers.Type {
	case SUBSCRIBE:
		e.subscribe(event.Payload.(string), conn)
	case UNSUBSCRIBE:
		e.unsubscribe(event.Payload.(string), conn)
	case BROADCAST:
		payload := event.Payload.(map[string]interface{})
		e.send(payload["channel"].(string), payload["message"].(string), conn)
	}
}

func (e *EventManager) Disconnect(conn net.Conn) {
	for channelName, _ := range e.channels {
		runtime.Gosched()
		e.unsubscribe(channelName, conn)
	}
}

func (e *EventManager) subscribe(channelName string, conn net.Conn) {
	if e.channels[channelName] == nil {
		e.registerChannel(channelName)
	}

	e.channels[channelName].AddSubscriber(conn)
}

func (e *EventManager) unsubscribe(channelName string, conn net.Conn) {
	if e.channels[channelName] == nil {
		return
	}
	e.channels[channelName].RemoveSubscriber(conn)
	if e.channels[channelName].IsEmpty() {
		delete(e.channels, channelName)
	}
}

func (e *EventManager) send(channelName string, message string, conn net.Conn) {
	if e.channels[channelName] == nil {
		return
	}
	if e.channels[channelName].IsEmpty() || !e.channels[channelName].Contains(conn) {
		return
	}
	e.channels[channelName].Broadcast(message + "\n")
}
