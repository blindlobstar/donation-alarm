package cevents

import (
	"errors"
	"log"

	"github.com/blindlobstar/donation-alarm/backend/internal/events"
)

type EventBus struct {
	c               chan events.Event
	eventHandlerMap map[string]events.EventHandler
}

func New(c chan events.Event) EventBus {
	return EventBus{
		c:               c,
		eventHandlerMap: make(map[string]events.EventHandler),
	}
}

func (eb *EventBus) RegisterHandler(handler events.EventHandler, name string) error {
	if _, ok := eb.eventHandlerMap[name]; ok {
		return errors.New("event handler already exists")
	}
	eb.eventHandlerMap[name] = handler
	return nil
}

func (eb *EventBus) Publish(payload any, name string) error {
	eb.c <- events.Event{
		Payload: payload,
		Name:    name,
	}
	return nil
}

func (eb *EventBus) Run() {
	event := <-eb.c

	handler, ok := eb.eventHandlerMap[event.Name]
	if !ok {
		log.Printf("eventhandler not found. EventName: %s\n", event.Name)
		return
	}

	if err := handler.Handle(event.Payload); err != nil {
		log.Printf("error while handling event. Name: %s, Error: %v\n", event.Name, err)
	}
}
