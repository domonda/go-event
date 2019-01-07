package event

import (
	"reflect"
	"sync"

	"github.com/domonda/Domonda/pkg/wrap"
	"github.com/domonda/errors"
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
func (stream *Stream) Publish(event interface{}) <-chan error {
	stream.handlerMtx.RLock()
	defer stream.handlerMtx.RUnlock()

	typeHandlers := stream.eventTypeHandlers[reflect.TypeOf(event)]

	var errs errors.Collection
	var wg sync.WaitGroup
	wg.Add(len(typeHandlers) + len(stream.anyEventHandlers))

	handleEventAsync := func(handler Handler, event interface{}) {
		errs.Add(safelyHandleEvent(handler, event))
		wg.Done()
	}

	for _, handler := range typeHandlers {
		go handleEventAsync(handler, event)
	}
	for _, handler := range stream.anyEventHandlers {
		go handleEventAsync(handler, event)
	}

	errChan := make(chan error, 1)
	go func() {
		wg.Wait()
		errChan <- errs.Combine()
	}()

	return errChan
}

func (stream *Stream) PublishAwait(event interface{}) (err error) {
	return <-stream.Publish(event)
}

func safelyHandleEvent(handler Handler, event interface{}) (err error) {
	defer wrap.RecoverPanicAsResultError(&err, "safelyHandleEvent")
	return handler.HandleEvent(event)
}
