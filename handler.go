package event

import (
	"reflect"

	"github.com/domonda/errors"
)

type Handler interface {
	HandleEvent(event interface{})
}

// HandlerFunc implements Handler for a function pointer
type HandlerFunc func(event interface{})

func (f HandlerFunc) HandleEvent(event interface{}) { f(event) }

func RepublishHandler(publisher Publisher) Handler {
	return HandlerFunc(func(event interface{}) {
		publisher.Publish(event)
	})
}

func ChanHandler(handlerChan chan<- interface{}) Handler {
	return HandlerFunc(func(event interface{}) {
		handlerChan <- event
	})
}

func TypeHandler(handlerFuncOrChan interface{}) (eventType reflect.Type, handler Handler) {
	handlerVal := reflect.ValueOf(handlerFuncOrChan)
	handlerType := handlerVal.Type()

	switch handlerType.Kind() {
	case reflect.Func:
		if handlerType.NumIn() != 1 {
			panic(errors.Errorf("event handler function must have 1 argument, but has %d", handlerType.NumIn()))
		}
		if handlerType.NumOut() > 0 {
			panic(errors.Errorf("event handler function must not have results, but has %d", handlerType.NumOut()))
		}
		eventType = handlerType.In(0)
		handler = HandlerFunc(func(event interface{}) {
			handlerVal.Call([]reflect.Value{reflect.ValueOf(event)})
		})

	case reflect.Chan:
		eventType = handlerType.Elem()
		handler = HandlerFunc(func(event interface{}) {
			handlerVal.Send(reflect.ValueOf(event))
		})

	default:
		panic(errors.Errorf("event handler is not a function or channel but '%T'", handlerFuncOrChan))
	}

	return eventType, handler
}
