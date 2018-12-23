package event

import "reflect"

type SyncStream struct {
	subscribable
}

func NewSyncStream(subscribeTo ...Subscribable) *SyncStream {
	stream := new(SyncStream)
	for _, source := range subscribeTo {
		source.Subscribe(RepublishHandler(stream))
	}
	return stream
}

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
