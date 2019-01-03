package event

import (
	"reflect"

	"github.com/domonda/Domonda/pkg/wrap"
)

// Stream is an event stream that implements Publisher and Subscribable
// for publishing events and subscribing to them.
// Every Handler.HandleEvent method call is started as a separate Go routine,
// meaning that event handling will be run unordered and in parallel.
//
// Use SyncStream if synchronous, ordered handling of events is needed.
//
// Stream is threadsafe.
type Stream struct {
	subscribable
}

// NewStream returns a new Stream with optional RepublishHandler
// subscriptions to the passed subscribeTo Subscribable implementations.
func NewStream(subscribeTo ...Subscribable) *Stream {
	stream := new(Stream)
	for _, source := range subscribeTo {
		source.Subscribe(RepublishHandler(stream))
	}
	return stream
}

// Publish calls Handler.HandleEvent(event) for all subscribed event handlers.
// Every Handler.HandleEvent method call is started as a separate Go routine,
// meaning that event handling will be run unordered and in parallel.
//
// Use SyncStream if synchronous, ordered handling of events is needed.
func (stream *Stream) Publish(event interface{}) {
	stream.handlerMtx.RLock()
	defer stream.handlerMtx.RUnlock()

	for _, handler := range stream.eventTypeHandlers[reflect.TypeOf(event)] {
		go safelyHandleEvent(handler, event)
	}
	for _, handler := range stream.anyEventHandlers {
		go safelyHandleEvent(handler, event)
	}
}

func safelyHandleEvent(handler Handler, event interface{}) {
	defer wrap.RecoverAndLogPanic("safelyHandleEvent")
	handler.HandleEvent(event)
}
