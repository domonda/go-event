package event

import (
	"reflect"
	"sync"
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

	typeHandlers := stream.eventTypeHandlers[reflect.TypeOf(event)]

	for _, handler := range typeHandlers {
		go safelyHandleEvent(handler, event)
	}
	for _, handler := range stream.anyEventHandlers {
		go safelyHandleEvent(handler, event)
	}
}

// PublishAsync publishes an event asynchronousely
// using one or more go routines.
// Exactly one error or nil will be written to
// the returned channel when the event has been
// handled by the subsribed handlers.
// The error can be a combination of multiple
// errors from multiple event handlers.
func (stream *Stream) PublishAsync(event interface{}) <-chan error {
	stream.handlerMtx.RLock()
	defer stream.handlerMtx.RUnlock()

	typeHandlers := stream.eventTypeHandlers[reflect.TypeOf(event)]
	var (
		errs    []error
		errsMtx sync.Mutex
		wg      sync.WaitGroup
	)
	wg.Add(len(typeHandlers) + len(stream.anyEventHandlers))

	handleEventAsync := func(handler Handler, event interface{}) {
		err := safelyHandleEvent(handler, event)
		if err != nil {
			errsMtx.Lock()
			errs = append(errs, err)
			errsMtx.Unlock()
		}
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
		errChan <- combineErrors(errs)
	}()

	return errChan
}

// PublishAwait publishes an event and waits
// for all handlers to return an error or nil.
// The error can be a combination of multiple
// errors from multiple event handlers.
func (stream *Stream) PublishAwait(event interface{}) error {
	return <-stream.PublishAsync(event)
}

func safelyHandleEvent(handler Handler, event interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = asError(r)
		}
	}()

	return handler.HandleEvent(event)
}
