package event

import (
	"reflect"
	"sync"
)

// Subscribable has a Subscribe method
type Subscribable interface {
	// Subscribe an event Handler to the passed eventTypes,
	// or to all events when no eventTypes are passed.
	Subscribe(handler Handler, eventTypes ...reflect.Type)
}

// SubscribableFunc implements Subscribable for a function pointer
type SubscribableFunc func(handler Handler, eventTypes ...reflect.Type)

func (f SubscribableFunc) Subscribe(handler Handler, eventTypes ...reflect.Type) {
	f(handler, eventTypes...)
}

// Subscription returns a Subscribable that passes along events from
// the eventSource Subscribable transformed with the passed transformFunc.
// See NewTransformer for how a Transformer is created from transformFunc.
func Subscription(eventSource Subscribable, transformFunc interface{}) Subscribable {
	transformer := NewTransformer(transformFunc)

	return SubscribableFunc(func(handler Handler, eventTypes ...reflect.Type) {
		// When subscribing only for certain eventTypes,
		// then check if the transformer returns one of those types,
		// if not, return without subscribing
		if len(eventTypes) > 0 && !containsType(eventTypes, transformer.ResultEventType()) {
			return
		}

		eventSource.Subscribe(
			HandlerFunc(func(event interface{}) error {
				transformed, useResult, err := transformer.TransformEvent(event)
				if !useResult || err != nil {
					return err
				}
				return handler.HandleEvent(transformed)
			}),
			transformer.SourceEventType(),
		)
	})
}

func containsType(types []reflect.Type, searched reflect.Type) bool {
	for _, t := range types {
		if t == searched {
			return true
		}
	}
	return false
}

// subscribable is the internal basis implementation
// of the Subscribable interface used by Stream and SyncStream
type subscribable struct {
	handlerMtx        sync.RWMutex
	eventTypeHandlers map[reflect.Type][]Handler
	anyEventHandlers  []Handler
}

// Subscribe an event Handler to the passed eventTypes,
// or to all events when no eventTypes are passed.
// This method implements the Subscribable interface.
func (s *subscribable) Subscribe(handler Handler, eventTypes ...reflect.Type) {
	s.handlerMtx.Lock()
	defer s.handlerMtx.Unlock()

	if len(eventTypes) == 0 {
		s.anyEventHandlers = append(s.anyEventHandlers, handler)
	} else {
		for _, eventType := range eventTypes {
			if s.eventTypeHandlers == nil {
				s.eventTypeHandlers = map[reflect.Type][]Handler{
					eventType: []Handler{handler},
				}
			} else {
				s.eventTypeHandlers[eventType] = append(s.eventTypeHandlers[eventType], handler)
			}
		}
	}
}

// SubscribeReflect uses reflection to wrap the passed eventHandler
// as Handler and subscribes it to a specific event type
// if the eventHandler is type specific,
// or to all events when no type can be derived from eventHandler.
//
// If eventHandler implements Handler,
// it will be subscribed to all event types.
//
// If eventHandler implements Publisher,
// it will be subscribed as RepublishHandler to all event types.
//
// If eventHandler is a function of type func(interface{}),
// it will be subscribed as HandlerFunc to all event types.
//
// If eventHandler is a channel of type chan interface{},
// it will be subscribed as ChanHandler to all event types.
//
// If eventHandler is not any of the above,
// it will be subscribed as TypeHandler for the event type
// returned by TypeHandler.
func (s *subscribable) SubscribeReflect(eventHandler interface{}) {
	switch x := eventHandler.(type) {
	case nil:
		panic("eventHandler is nil")

	case Handler:
		s.Subscribe(x)

	case Publisher:
		s.Subscribe(RepublishHandler(x))

	case func(interface{}) error:
		s.Subscribe(HandlerFunc(x))

	case func(interface{}):
		s.Subscribe(HandlerFuncNoError(x))

	case chan interface{}:
		s.Subscribe(ChanHandler(x))

	default:
		eventType, handler := TypeHandler(eventHandler)
		s.Subscribe(handler, eventType)
	}
}
