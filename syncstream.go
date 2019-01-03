package event

import "reflect"

// SyncStream is an event stream that implements Publisher and Subscribable
// for publishing events and subscribing to them.
// Handler.HandleEvent method calls are done synchronously from Publish,
// meaning that Publish will only return after all event handlers have been called
// in the order they have been subscribed.
//
// Use Stream instead if the events should be published
// asynchronously in parallel Go routines.
//
// SyncStream is threadsafe.
type SyncStream struct {
	subscribable
}

// NewSyncStream returns a new Stream with optional RepublishHandler
// subscriptions to the passed subscribeTo Subscribable implementations.
func NewSyncStream(subscribeTo ...Subscribable) *SyncStream {
	stream := new(SyncStream)
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
func (stream *SyncStream) Publish(event interface{}) {
	stream.handlerMtx.RLock()
	defer stream.handlerMtx.RUnlock()

	for _, handler := range stream.eventTypeHandlers[reflect.TypeOf(event)] {
		safelyHandleEvent(handler, event)
	}
	for _, handler := range stream.anyEventHandlers {
		safelyHandleEvent(handler, event)
	}
}
