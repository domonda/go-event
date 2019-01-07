package event

import (
	"reflect"
)

// SerialStream is an event stream that implements Publisher and Subscribable
// for publishing events and subscribing to them.
// Handler.HandleEvent method calls are done synchronously from Publish,
// meaning that Publish will only return after all event handlers have been called
// in the order they have been subscribed.
//
// Use Stream instead if the events should be published
// asynchronously in parallel Go routines.
//
// SerialStream is threadsafe.
type SerialStream struct {
	subscribable
}

// NewSerialStream returns a new Stream with optional RepublishHandler
// subscriptions to the passed subscribeTo Subscribable implementations.
func NewSerialStream(subscribeTo ...Subscribable) *SerialStream {
	stream := new(SerialStream)
	for _, source := range subscribeTo {
		source.Subscribe(RepublishHandler(stream))
	}
	return stream
}

// Publish calls Handler.HandleEvent(event) for all subscribed event handlers.
// First all type specific handlers are called in the order
// they have been subscribed for the type of the event.
// Then all non type specific handlers are called in the order
// they have been subscribed.
//
// Use Stream instead if the events should be published
// asynchronously in parallel Go routines.
func (stream *SerialStream) Publish(event interface{}) {
	stream.PublishAwait(event)
}

func (stream *SerialStream) PublishAsync(event interface{}) <-chan error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- stream.PublishAwait(event)
	}()
	return errChan
}

func (stream *SerialStream) PublishAwait(event interface{}) error {
	stream.handlerMtx.RLock()
	defer stream.handlerMtx.RUnlock()

	for _, handler := range stream.eventTypeHandlers[reflect.TypeOf(event)] {
		err := safelyHandleEvent(handler, event)
		if err != nil {
			return err
		}
	}
	for _, handler := range stream.anyEventHandlers {
		err := safelyHandleEvent(handler, event)
		if err != nil {
			return err
		}
	}

	return nil
}
