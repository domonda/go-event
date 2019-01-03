package event

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/domonda/errors"
)

// Handler has a method to handle an event of type interface{}
type Handler interface {
	// HandleEvent handles an event
	HandleEvent(event interface{})
}

// HandlerFunc implements Handler for a function pointer
type HandlerFunc func(event interface{})

func (f HandlerFunc) HandleEvent(event interface{}) { f(event) }

// RepublishHandler wraps a Publisher as a Handler
// by calling publisher.Publish(event) in every
// HandleEvent method call.
func RepublishHandler(publisher Publisher) Handler {
	return HandlerFunc(func(event interface{}) {
		publisher.Publish(event)
	})
}

// ChanHandler wraps a channel as an event Handler.
// HandleEvent calls will write the passed events to the channel.
func ChanHandler(handlerChan chan<- interface{}) Handler {
	return HandlerFunc(func(event interface{}) {
		handlerChan <- event
	})
}

// WriteJSONHandler returns a Handler that pretty prints
// every event as JSON followed by a newline to writer.
// Errors from encoding and writing create a panic.
func WriteJSONHandler(writer io.Writer) Handler {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	return HandlerFunc(func(event interface{}) {
		err := encoder.Encode(event)
		if err != nil {
			panic(err)
		}
		_, err = writer.Write([]byte{'\n'})
		if err != nil {
			panic(err)
		}
	})
}

// TypeHandler
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
