package prago

import (
	"errors"
	"fmt"
)

var (
	EventsErrorNotFound = errors.New("No event listener found")
	EventsErrorMultiple = errors.New("Multiple event listeners found")
)

type Events struct {
	events map[string][]Event
}

func NewEvents() *Events {
	return &Events{
		make(map[string][]Event),
	}
}

type Event func(data interface{}) (interface{}, error)

func (e *Events) Listen(eventType string, fn Event) {
	fmt.Println("listen", eventType)
	e.events[eventType] = append(e.events[eventType], fn)
}

func (e *Events) Send(eventType string, data interface{}) (err error) {
	fmt.Println(eventType, data, e.events[eventType])
	for _, v := range e.events[eventType] {
		_, err = v(data)
		if err != nil {
			return
		}
	}
	return
}

func (e *Events) Get(eventType string, data interface{}) (interface{}, error) {
	events := e.events[eventType]
	if len(events) == 1 {
		return events[0](data)
	} else {
		if len(events) == 0 {
			return nil, EventsErrorNotFound
		} else {
			return nil, EventsErrorMultiple
		}
	}
}
