package event

import (
	"reflect"
	"sync"
)

type Subscribable interface {
	Subscribe(handler Handler, eventTypes ...reflect.Type)
}

// SubscribableFunc implements Source for a function pointer
type SubscribableFunc func(handler Handler, eventTypes ...reflect.Type)

func (f SubscribableFunc) Subscribe(handler Handler, eventTypes ...reflect.Type) {
	f(handler, eventTypes...)
}

func Subscription(eventSource Subscribable, transformFunc interface{}) Subscribable {
	transformer := NewTransformer(transformFunc)

	return SubscribableFunc(func(handler Handler, eventTypes ...reflect.Type) {
		if len(eventTypes) > 0 {
			resultEventType := transformer.ResultEventType()
			found := false
			for _, eventType := range eventTypes {
				if eventType == resultEventType {
					found = true
					break
				}
			}
			if !found {
				return
			}
		}

		eventSource.Subscribe(
			HandlerFunc(func(event interface{}) {
				transformed, ok := transformer.TransformEvent(event)
				if ok {
					handler.HandleEvent(transformed)
				}
			}),
			transformer.SourceEventType(),
		)
	})
}

type subscribable struct {
	handlerMtx        sync.RWMutex
	eventTypeHandlers map[reflect.Type][]Handler
	anyEventHandlers  []Handler
}

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

func (s *subscribable) SubscribeReflect(eventHandler interface{}) {
	switch x := eventHandler.(type) {
	case nil:
		panic("eventHandler is nil")

	case Handler:
		s.Subscribe(x)

	case Publisher:
		s.Subscribe(RepublishHandler(x))

	case func(interface{}):
		s.Subscribe(HandlerFunc(x))

	case chan interface{}:
		s.Subscribe(ChanHandler(x))

	default:
		eventType, handler := TypeHandler(eventHandler)
		s.Subscribe(handler, eventType)
	}
}
