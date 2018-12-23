package event

import (
	"reflect"

	"github.com/domonda/Domonda/pkg/wrap"
)

type Stream struct {
	subscribable
}

func NewStream(subscribeTo ...Subscribable) *Stream {
	stream := new(Stream)
	for _, source := range subscribeTo {
		source.Subscribe(RepublishHandler(stream))
	}
	return stream
}

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
